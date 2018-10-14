// NEC IR transsion protocol in go

package nec

import (
	"time"
)

/*
	examples of NEC protocol

	classic:
		addr = 0
		cmd  = 1
		data = addr + ^addr + cmd + ^cmd
			=> 00 FF 80 7F

	      9 ms      | 4.5 ms |         addr          |                 ^addr                 |          cmd            |                 ^cmd                | end
	1111111111111111 00000000 10 10 10 10 10 10 10 10 1000 1000 1000 1000 1000 1000 1000 1000 1000 10 10 10 10 10 10 10 10 1000 1000 1000 1000 1000 1000 1000 10000000...
	                           0  0  0  0  0  0  0  0   1    1    1    1    1    1    1    1    1   0  0  0  0  0  0  0  0    1    1    1    1    1    1    1

	extended:
		addr = 61184
		cmd = 3
		data = low_addr + high_addr + cmd + ^cmd
			=> 00 F7 C0 3F

	      9 ms      | 4.5 ms |        low addr       |               high addr             |          cmd              |                 ^cmd              | end
	1111111111111111 00000000 10 10 10 10 10 10 10 10 1000 1000 1000 1000 10 1000 1000 1000 1000 1000 10 10 10 10 10 10 10 10 1000 1000 1000 1000 1000 1000 10000000...
	                           0  0  0  0  0  0  0  0   1    1    1    1   0   1    1    1    1    1   0  0  0  0  0  0  0  0   1    1    1    1    1    1

	repeat code:

	      9 ms      |2.25|
	1111111111111111 0000 1000...
*/

// Base time unit in which signal level is unchanged
const Tick = time.Millisecond * 9 / 16 // 562.5 us
// Length of signal in Ticks
const SigLength = 192 // 192 * Tick = 108 ms

var repeatSignal = NewRepeatSignal()

type Signal []bool

func (s Signal) String() string {
	l := len([]bool(s))
	var runes = make([]rune, l)
	for i, bit := range []bool(s) {
		if bit {
			runes[i] = '1'
		} else {
			runes[i] = '0'
		}
	}
	return string(runes)
}

func newSignalBegin() Signal { // 1111111111111111 00000000
	signal := make(Signal, 16+8, SigLength)
	for i := 0; i < 16; i++ {
		signal[i] = true
	}
	return signal
}

func (s *Signal) appendByte(b byte) {
	for i := uint(0); i < 8; i++ {
		if b>>i&1 == 0 { // lsb first
			*s = append(*s, true, false) // logic 0 => |-_|
		} else {
			*s = append(*s, true, false, false, false) // logic 1 => |-___|
		}
	}
}

func (s *Signal) enclose() Signal {
	*s = append(*s, true)
	*s = (*s)[0:SigLength]
	return *s
}

// Transmits signal once
func (s Signal) Transmit(f func(bool)) {
	ticker := time.NewTicker(Tick)
	for _, v := range []bool(s) {
		<-ticker.C
		f(v)
	}
	ticker.Stop()
}

// Transmits same signal n times
func (s Signal) TransmitTimes(f func(bool), n int) {
	ticker := time.NewTicker(Tick)
	for i := 0; i < n; i++ {
		for _, v := range []bool(s) {
			<-ticker.C
			f(v)
		}
	}
	ticker.Stop()
}

// Transmits signal once with following repeat code signal n times
func (s Signal) TransmitRepeat(f func(bool), n int) {
	ticker := time.NewTicker(Tick)
	for _, v := range []bool(s) {
		<-ticker.C
		f(v)
	}
	for i := 0; i < n; i++ {
		for _, v := range []bool(repeatSignal) {
			<-ticker.C
			f(v)
		}
	}
	ticker.Stop()
}

func NewSignal(code uint32) Signal {
	s := newSignalBegin()
	for i := uint(0); i < 32; i++ { // msb first
		if code<<i&(1<<31) == 0 {
			s = append(s, true, false) // logic 0 => |-_|
		} else {
			s = append(s, true, false, false, false) // logic 1 => |-___|
		}
	}
	return s.enclose()
}

func NewRepeatSignal() Signal { // 1111111111111111 0000 1000...
	signal := newSignalBegin()[0 : 16+4]
	return signal.enclose()
}

// Encodes address and comand to nec signal
func Encode(addr uint8, cmd uint8) Signal {
	signal := newSignalBegin()
	signal.appendByte(addr)
	signal.appendByte(^addr)
	signal.appendByte(cmd)
	signal.appendByte(^cmd)
	signal.enclose()
	return signal
}

// Encodes extended adress and command to nec signal
func EncodeExt(addr uint16, cmd uint8) Signal {
	signal := newSignalBegin()
	signal.appendByte(byte(addr)) // low byte first
	signal.appendByte(byte(addr >> 8))
	signal.appendByte(cmd)
	signal.appendByte(^cmd)
	signal.enclose()
	return signal
}
