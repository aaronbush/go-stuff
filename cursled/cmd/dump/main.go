package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aaronbush/go-stuff/cursled/frame"
	"github.com/gookit/color"
)

type ScreenKey struct {
	Row    uint8
	Column uint8
}

func main() {
	file, err := os.Open("test.data")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	ledInfo := frame.LEDInfo{}
	ledInfos := make(map[ScreenKey]frame.LEDInfo)

	for {
		if err := binary.Read(file, binary.BigEndian, &ledInfo); err != nil {
			log.Fatal("error reading in the led info", err)
			return
		}
		ledInfos[ScreenKey{ledInfo.Row, ledInfo.Column}] = ledInfo
		fmt.Printf("%+v\n", ledInfo)
		// drawTable(ledInfos)
	}
}

func drawTable(leds map[ScreenKey]frame.LEDInfo, numRows uint8, numColumns uint8) {
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
