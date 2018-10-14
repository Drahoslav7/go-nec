package nec

import (
	"fmt"
	"testing"
)

func ExampleEncode() {
	fmt.Println(Encode(0, 65))         // 44 key remote - ON
	fmt.Println(NewSignal(0x00FF827D)) // 44 key remote - ON

	fmt.Println(EncodeExt(61184, 3))   // 24 key remote - ON
	fmt.Println(NewSignal(0x00F7C03F)) // 24 key remote - ON
	fmt.Println(NewRepeatSignal())
}

func ExampleSignal_Transmit() {
	connect := func(v bool) {
		if v {
			fmt.Print(1) // imagine writing to gpio pins here
		} else {
			fmt.Print(0)
		}
	}
	sig := EncodeExt(61184, 3)
	sig.Transmit(connect)
	fmt.Println()
	sig.TransmitTimes(connect, 4)
	fmt.Println()
	sig.TransmitRepeat(connect, 4)
	fmt.Println()
}

func TestEncode(t *testing.T) {
	s := Encode(0, 1)
	if s.String()[:128] != "11111111111111110000000010101010101010101000100010001000100010001000100010001010101010101010100010001000100010001000100010000000" {
		t.Error("Signal is not Encoded correctly - nec")
	}
}

func TestEncodeExt(t *testing.T) {
	s := EncodeExt(61184, 3)
	if s.String()[:128] != "11111111111111110000000010101010101010101000100010001000101000100010001000100010101010101010101000100010001000100010001000000000" {
		t.Error("Signal is note Encoded correctly - Extended nec")
	}
}
