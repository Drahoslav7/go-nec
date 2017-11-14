package nec

import (
	"fmt"
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
