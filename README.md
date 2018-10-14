# go-nec
Implementation of IR transmition NEC protocol in golang


## usage
```go
package main

import (
        "os"
        "flag"
        "github.com/stianeikeland/go-rpio"
        "github.com/drahoslav7/go-nec"
)

var (
        addr uint = 61184 // 24 key ir remote control
        cmd uint = 3 // ON
)

func init() {
        flag.UintVar(&cmd, "c", cmd, "command to send (0-23)")
        flag.Parse()
}

func main() {
        // init gpio
        err := rpio.Open()
        if err != nil {
                os.Exit(1)
        }
        defer rpio.Close()

        pin := rpio.Pin(2)
        pin.Mode(rpio.Output)
        pin.Write(rpio.Low)

        // transmition funcion
        toLED := func(v bool) {
                if v {
                        pin.Write(rpio.High)
                } else {
                        pin.Write(rpio.Low)
                }
        }
        nec.EncodeExt(uint16(addr), uint8(cmd)).TransmitTimes(toLED, 3)
}

```


## examples of NEC protocol

#### classic:

`addr = 0`
`cmd  = 1`

`data = addr + ^addr + cmd + ^cmd`
=> `00 FF 80 7F`
```
      9 ms      | 4.5 ms |         addr          |                 ^addr                 |          cmd            |                 ^cmd                | end
1111111111111111 00000000 10 10 10 10 10 10 10 10 1000 1000 1000 1000 1000 1000 1000 1000 1000 10 10 10 10 10 10 10 10 1000 1000 1000 1000 1000 1000 1000 10000000...
                           0  0  0  0  0  0  0  0   1    1    1    1    1    1    1    1    1   0  0  0  0  0  0  0  0    1    1    1    1    1    1    1
```
`nec.Encode(0, 1)` or `nec.NewSignal(0x00FF807F)`

#### extended:

`addr = 61184`
`cmd = 3`

`data = low_addr + high_addr + cmd + ^cmd`
=> `00 F7 C0 3F`
```    
      9 ms      | 4.5 ms |        low addr       |               high addr             |          cmd              |                 ^cmd              | end
1111111111111111 00000000 10 10 10 10 10 10 10 10 1000 1000 1000 1000 10 1000 1000 1000 1000 1000 10 10 10 10 10 10 10 10 1000 1000 1000 1000 1000 1000 10000000...
                           0  0  0  0  0  0  0  0   1    1    1    1   0   1    1    1    1    1   0  0  0  0  0  0  0  0   1    1    1    1    1    1
```
`nec.EncodeExt(61184, 3)` or `nec.NewSignal(0x00F7C03F)`

#### repeat code:
```
      9 ms      |2.25|
1111111111111111 0000 1000...
```
`nec.NewRepeatSignal()`
