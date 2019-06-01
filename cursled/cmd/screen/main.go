package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/gookit/color"
)

const StartSentinel uint32 = 0xDEADBEEF

type Header struct {
	NumLEDs uint16
}

type ScreenKey struct {
	Row    uint8
	Column uint8
}

type LEDInfo struct {
	Row        uint8
	Column     uint8
	Red        uint8
	Green      uint8
	Blue       uint8
	Brightness uint8
}

func main() {
	maxLeds := uint16(1000)
	maxRows := uint8(40)
	numRows := uint8(0)
	maxColumns := uint8(20)
	numColumns := uint8(0)

	// i, _ := os.Open(os.Args[1])
	// defer i.Close()

	b := makeLedsData()
	i := bytes.NewReader(b)

	ready := false
	dataIn := make([]byte, 4)

	for {
		if !ready {
			_, err := i.Read(dataIn)
			if err == io.EOF {
				return
			} else if err != nil {
				log.Fatal("error reading for start sentinel", err)
			}
			if binary.BigEndian.Uint32(dataIn) == StartSentinel {
				ready = true
			}
		}

		// we are ready...
		header := Header{}

		if err := binary.Read(i, binary.BigEndian, &header); err != nil {
			log.Fatal("error reading in the header", err)
			return
		}
		header.NumLEDs = min(header.NumLEDs, maxLeds)

		ledInfo := LEDInfo{}
		log.Printf("Reading in info for %d LEDs\n", header.NumLEDs)

		ledInfos := make(map[ScreenKey]LEDInfo)

		for n := uint16(0); n < header.NumLEDs; n++ {
			log.Println("reading in info for led #", n)
			if err := binary.Read(i, binary.BigEndian, &ledInfo); err != nil {
				log.Fatal("error reading in the led info", err)
				return
			}
			if ledInfo.Column > maxColumns || ledInfo.Row > maxRows {
				continue // discard out of range leds
			}
			ledInfos[ScreenKey{ledInfo.Row, ledInfo.Column}] = ledInfo

			numColumns = max8(ledInfo.Column, numColumns)
			numRows = max8(ledInfo.Row, numRows)
		}
		numColumns = min8(numColumns, maxColumns)
		numRows = min8(numRows, maxRows)
		drawTable(ledInfos, numRows, numColumns)
		ready = false
	}
}

func drawTable(leds map[ScreenKey]LEDInfo, numRows uint8, numColumns uint8) {
	log.Printf("%dx%d -> %v\n", numRows, numColumns, leds)
	black := color.BgBlack.Sprint("  ")
	for row := uint8(1); row <= numRows; row++ {
		fmt.Printf("%02d:", row)
		var sb strings.Builder
		for column := uint8(1); column <= numColumns; column++ {
			if led, ok := leds[ScreenKey{row, column}]; ok {
				c := color.RGB(led.Red, led.Green, led.Blue, true)
				sb.WriteString(c.Sprintf("  "))
			} else {
				sb.WriteString(black)
			}
			fmt.Printf(sb.String())
			sb.Reset()
		}
		fmt.Printf(":%02d\n", row)
	}
}

func max(x, y uint16) uint16 {
	if x > y {
		return x
	}
	return y
}
func max8(x, y uint8) uint8 {
	return uint8(max(uint16(x), uint16(y)))
}

func min(x, y uint16) uint16 {
	if max(x, y) == x {
		return y
	}
	return x
}

func min8(x, y uint8) uint8 {
	return uint8(min(uint16(x), uint16(y)))
}

func filter(leds []LEDInfo, test func(LEDInfo) bool) (ret []LEDInfo) {
	for _, l := range leds {
		if test(l) {
			ret = append(ret, l)
		}
	}
	return
}

func makeLedsData() []byte {
	return []byte{0xDE, 0xAD, 0xBE, 0xEF, // start sentinel
		0x00, 0x06, // numLEDs

		0x01, 0x01, // row, column
		0xff, 0x00, 0x00, // RGB
		0xFF, // Brightness

		0x01, 0x02, // row, column
		0x00, 0xff, 0xff, // RGB
		0xFF, // Brightness

		0x01, 0x03, // row, column
		0xff, 0xff, 0xff, // RGB
		0xFF, // Brightness

		// row 2
		0x02, 0x01, // row, column
		0xff, 0x00, 0x00, // RGB
		0xFF, // Brightness

		0x02, 0x02, // row, column
		0x00, 0xff, 0xff, // RGB
		0xFF, // Brightness

		0x02, 0x03, // row, column
		0xff, 0xff, 0xff, // RGB
		0xFF, // Brightness
	}
}
