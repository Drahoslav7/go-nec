// NEC IR transsion protocol in go

package nec

import (
	"fmt"
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

const Tick = time.Millisecond * 9/16 // 562.5 us
const MaxLength = 192 // 192 * Tick = 108 ms

var RepeatSignal = NewRepeatSignal()

type Signal []bool

func (s Signal) String () string {
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
	signal := make(Signal, 16+8, MaxLength)
	for i := 0; i < 16; i++ {
		signal[i] = true
	}
	return signal
}

func (s* Signal) appendByte(b byte) {
	for i := uint(0); i < 8; i++ {
		if b >> i & 1 == 0 { // lsb first
			*s = append(*s, true, false) // logic 0 => |-_|
		} else {
			*s = append(*s, true, false, false, false)	// logic 1 => |-___|
		}
	}
}

func (s* Signal) enclose() Signal {
	*s = append(*s, true)
	*s = (*s)[0:MaxLength]
	return *s
}

/**
 * Transmits signal once
 */
func (s Signal) Transmit(f func(bool)) {
	ticker := time.NewTicker(Tick)
	for _, v := range []bool(s) {
		f(v)
		<-ticker.C
	}
	ticker.Stop()
}

/**
 * Transmits same signal n times
 */
func (s Signal) TransmitTimes(f func(bool), n int) {
	ticker := time.NewTicker(Tick)
	for i := 0; i < n; i++ {
		for _, v := range []bool(s) {
			f(v)
			<-ticker.C
		}
	}
	ticker.Stop()
}

/**
 * Transmit signal once with following repeat code signal n times
 */
func (s Signal) TransmitRepeat(f func(bool), n int) {
	ticker := time.NewTicker(Tick)
	for _, v := range []bool(s) {
		f(v)
		<-ticker.C
	}
	for i := 0; i < n; i++ {
		for _, v := range []bool(RepeatSignal) {
			f(v)
			<-ticker.C
		}
	}
	ticker.Stop()
}

func NewSignal(code uint32) Signal {
	s := newSignalBegin()
	for i := uint(0); i < 32; i++ { // msb first
		if code << i & (1 << 31) == 0 {
			s = append(s, true, false) // logic 0 => |-_|
		} else {
			s = append(s, true, false, false, false)	// logic 1 => |-___|
		}
	}
	return s.enclose()
}

func NewRepeatSignal() Signal { // 1111111111111111 0000 1000...
	signal := newSignalBegin()[0:16+4]
	return signal.enclose()
}

func Encode(addr uint8, cmd uint8) Signal {
	signal := newSignalBegin()
	signal.appendByte(addr)
	signal.appendByte(^addr)
	signal.appendByte(cmd)
	signal.appendByte(^cmd)
	signal.enclose()
	return signal
}

func EncodeExt(addr uint16, cmd uint8) Signal {
	signal := newSignalBegin()
	signal.appendByte(byte(addr)) // low byte first
	signal.appendByte(byte(addr >> 8))
	signal.appendByte(cmd)
	signal.appendByte(^cmd)
	signal.enclose()
	return signal
}

func Example() {
	fmt.Println(Encode(0, 65))         // 44 key remote - ON
	fmt.Println(NewSignal(0x00FF827D)) // 44 key remote - ON

	fmt.Println(EncodeExt(61184, 3))   // 24 key remote - ON
	fmt.Println(NewSignal(0x00F7C03F)) // 24 key remote - ON
	fmt.Println(NewRepeatSignal())

	connect := func(v bool){
		if v {
			fmt.Print(1) // imagine witing to gpio pins here
		} else {
			fmt.Print(0)
		}
	}
	s := EncodeExt(61184, 3)
	s.Transmit(connect)
	fmt.Println()
	s.TransmitTimes(connect, 4)
	fmt.Println()
	s.TransmitRepeat(connect, 4)
	fmt.Println()
}