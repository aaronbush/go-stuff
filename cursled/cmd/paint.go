// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log"
	"os"
	"time"

	"github.com/aaronbush/go-stuff/cursled/frame"
	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/spf13/cobra"
)

type SquareInfo struct {
	GridCord  GridCord
	Vector2   rl.Vector2
	CreatedAt time.Time
	Color     rl.Color
}

type GridCord struct {
	Row    uint8
	Column uint8
}

var (
	numRows       int32
	numColumns    int32
	spacing       int32
	spacingFloat  float32
	maxBrightness float32
	decayTime     time.Duration
	windowHeight  int32
	windowWidth   int32
)

// paintCmd represents the paint command
var paintCmd = &cobra.Command{
	Use:   "paint",
	Short: "Start the paint UI",
	Long:  `Start the paint-like UI which allows drawing directly on the LED display`,
	RunE:  paint,
}

func init() {
	rootCmd.AddCommand(paintCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// paintCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// paintCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	paintCmd.Flags().Int32VarP(&numRows, "rows", "r", 40, "number of rows")
	paintCmd.Flags().Int32VarP(&numColumns, "columns", "c", 20, "number of columns")
	paintCmd.Flags().Int32VarP(&spacing, "spacing", "s", 20, "cell spacing")
	paintCmd.Flags().Float32VarP(&maxBrightness, "brightness", "b", 50, "max brightness")
	paintCmd.Flags().DurationVarP(&decayTime, "decayTime", "t", 3*time.Second, "decay time (seconds)")
}

func paint(cmd *cobra.Command, args []string) error {
	windowHeight = numRows * spacing
	windowWidth = numColumns * spacing
	traceContents := make(map[GridCord]SquareInfo)
	drawnContents := make(map[GridCord]SquareInfo)
	spacingFloat = float32(spacing)

	trackMouse := false
	fadeMode := false
	logMode := true

	// Open a new file for writing only
	file, err := os.OpenFile(
		"test.data",
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0666,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	rl.InitWindow(windowWidth, windowHeight, "pixel drawing")

	rl.SetTargetFPS(30)

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		drawGrid()
		drawSquares(traceContents, fadeMode)
		drawSquares(drawnContents, fadeMode)

		if logMode {
			exportSquares(file, drawnContents, fadeMode)
		}

		if rl.IsKeyPressed(rl.KeyF) {
			fadeMode = !fadeMode
			log.Printf("Fade Mode: %t", fadeMode)
		}
		if rl.IsKeyPressed(rl.KeyT) {
			trackMouse = !trackMouse
			log.Printf("Track Mode: %t", trackMouse)
		}
		if rl.IsKeyPressed(rl.KeyL) {
			logMode = !logMode
			log.Printf("Log Mode: %t", logMode)
		}

		if trackMouse {
			mousePos := rl.GetMousePosition()
			squareInfo, err := squareFromCoord(mousePos)

			if err == nil {
				//fmt.Println(mousePos, " -> ", squareInfo)
				if rl.IsMouseButtonDown(rl.MouseLeftButton) {
					squareInfo.Color = rl.Red
					drawnContents[squareInfo.GridCord] = squareInfo
				} else {
					squareInfo.Color = rl.Blue
					traceContents[squareInfo.GridCord] = squareInfo
				}
			}
		}
		rl.EndDrawing()
	}

	rl.CloseWindow()
	return nil
}

func drawGrid() {
	for r := int32(0); r <= windowHeight; r += spacing {
		rl.DrawLine(0, r, windowWidth, r, rl.Black)
	}
	for c := int32(0); c <= windowWidth; c += spacing {
		rl.DrawLine(c, 0, c, windowHeight, rl.Black)
	}
}

func drawSquares(squareContets map[GridCord]SquareInfo, fadeMode bool) {
	for cord, square := range squareContets {
		if timeLeft := time.Now().Sub(square.CreatedAt); timeLeft < decayTime {
			// scale for alpha
			var alpha = float32(1.0)
			if fadeMode {
				alpha = 1.0 - float32(timeLeft.Nanoseconds())/float32(decayTime.Nanoseconds())
			}
			rl.DrawRectangleV(square.Vector2, rl.NewVector2(spacingFloat, spacingFloat), rl.Fade(square.Color, alpha))
		} else {
			delete(squareContets, cord)
		}
	}
}

func exportSquares(file *os.File, squares map[GridCord]SquareInfo, fadeMode bool) {
	ledInfo := frame.LEDInfo{}
	var binBuf bytes.Buffer
	for _, square := range squares {
		if timeLeft := time.Now().Sub(square.CreatedAt); timeLeft < decayTime {
			// scale for brightness
			var alpha = float32(1.0)
			if fadeMode {
				alpha = 1.0 - float32(timeLeft.Nanoseconds())/float32(decayTime.Nanoseconds())
			}
			// write to I/O
			ledInfo.Column = square.GridCord.Column
			ledInfo.Row = square.GridCord.Row
			ledInfo.Brightness = uint8(maxBrightness * alpha)
			alpha = alpha
			err := binary.Write(&binBuf, binary.BigEndian, ledInfo)
			if err != nil {
				panic(err)
			}
			_, err = file.Write(binBuf.Bytes())
			if err != nil {
				log.Println(binBuf)
				panic(err)
			}
		}
	}
}

func squareFromCoord(vec rl.Vector2) (SquareInfo, error) { // return top left of square
	x := int32(vec.X)
	y := int32(vec.Y)

	if x <= 0 || y <= 0 ||
		x >= windowWidth || y >= windowHeight {
		return SquareInfo{}, errors.New("Outside of screen bounds")
	}
	xPos := x / spacing
	yPos := y / spacing

	info := SquareInfo{
		GridCord:  GridCord{uint8(xPos), uint8(yPos)},
		Vector2:   rl.NewVector2(float32(xPos*spacing), float32(yPos*spacing)),
		CreatedAt: time.Now(),
	}
	return info, nil
}
