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
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/aaronbush/go-stuff/cursled/frame"
	rg "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/spf13/cobra"
)

type SquareInfo struct {
	GridCord  GridCord
	Origin    rl.Vector2
	CreatedAt time.Time
	Color     rl.Color
}

type GridCord struct {
	Row    uint8
	Column uint8
}

const (
	maxRGB = 255
)

var (
	numRows            int32
	numColumns         int32
	spacing            int32
	spacingFloat       float32
	maxBrightness      float32
	decayTime          time.Duration
	windowHeight       int32
	windowWidth        int32
	gridOrigin         rl.Vector2
	gridHeight         int32
	gridWidth          int32
	statusBarOrigin    rl.Vector2
	rightControlOrigin rl.Vector2
	stausBarHeight     int32 = 60
	rightControlWidth  int32 = 60
	gridColor                = rl.RayWhite
	decayMode                = false
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
	gridOrigin = rl.NewVector2(0, 0)
	gridHeight = numRows * spacing
	gridWidth = numColumns * spacing

	statusBarOrigin = rl.NewVector2(gridOrigin.X, gridOrigin.Y+float32(gridHeight))
	rightControlOrigin = rl.NewVector2(gridOrigin.X+float32(gridWidth), gridOrigin.Y)

	windowHeight = gridHeight + stausBarHeight
	windowWidth = gridWidth + stausBarHeight

	traceContents := make(map[GridCord]SquareInfo)
	drawnContents := make(map[GridCord]SquareInfo)
	spacingFloat = float32(spacing)

	redValue, greenValue, blueValue := new(int), new(int), new(int)

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
	rg.LoadGuiStyle("styles/solarized_light.style")

	rl.SetTargetFPS(30)

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.Blank)

		drawColor, decayOrigin := drawColorInputs(rightControlOrigin, redValue, greenValue, blueValue)
		decayMode, _ = drawDecaySettings(decayOrigin, &decayMode)

		drawSquares(traceContents, fadeMode)
		drawSquares(drawnContents, fadeMode)

		drawGrid() // after colors are drawn to keep grid lines

		statusText := fmt.Sprintf("fade:%t, track:%t, log:%t\nFPS: %.1f (%.03f)", fadeMode, trackMouse, logMode, rl.GetFPS(), rl.GetFrameTime())
		rl.DrawText(statusText, int32(statusBarOrigin.X+3), int32(statusBarOrigin.Y), 12, rl.Gray)

		if logMode {
			exportSquares(file, drawnContents, fadeMode)
		}

		if rl.IsKeyPressed(rl.KeyF) {
			fadeMode = !fadeMode
		}
		if rl.IsKeyPressed(rl.KeyT) {
			trackMouse = !trackMouse
		}
		if rl.IsKeyPressed(rl.KeyL) {
			logMode = !logMode
		}

		if trackMouse {
			mousePos := rl.GetMousePosition()
			squareInfo, err := squareFromCoord(mousePos)

			if err == nil {
				if rl.IsMouseButtonDown(rl.MouseRightButton) {
					squareInfo.Color = rl.Blank
					drawnContents[squareInfo.GridCord] = squareInfo
				} else if rl.IsMouseButtonDown(rl.MouseLeftButton) {
					squareInfo.Color = drawColor
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
	// draw row lines
	for rowNum, rowBegin := int32(0), gridOrigin; rowNum <= numRows; rowNum++ {
		rowEnd := rl.NewVector2(rowBegin.X+float32(gridWidth), rowBegin.Y)
		rl.DrawLineEx(rowBegin, rowEnd, 1.0, gridColor)
		rowBegin.Y += float32(spacing)
	}
	// draw column lines
	for colNum, colBegin := int32(0), gridOrigin; colNum <= numColumns; colNum++ {
		colEnd := rl.NewVector2(colBegin.X, colBegin.Y+float32(gridHeight))
		rl.DrawLineEx(colBegin, colEnd, 1.0, gridColor)
		colBegin.X += float32(spacing)
	}
}

func drawColorInputs(position rl.Vector2, red, green, blue *int) (rl.Color, rl.Vector2) {
	position.X += 5
	drawColorInput("red", red, position)
	position.Y += 45
	drawColorInput("green", green, position)
	position.Y += 45
	drawColorInput("blue", blue, position)
	position.Y += 45

	color := makeColor(*red, *green, *blue, 255)
	rl.DrawRectangleV(position, rl.NewVector2(spacingFloat, spacingFloat), color)
	return color, rl.NewVector2(position.X, position.Y+spacingFloat)
}

func drawColorInput(name string, colorValue *int, position rl.Vector2) {
	rg.Label(rl.NewRectangle(position.X, position.Y, 50, 20), name)
	color := rg.TextBox(rl.NewRectangle(position.X, position.Y+20, 50, 20), strconv.Itoa(*colorValue))
	*colorValue, _ = strconv.Atoi(color)
	if *colorValue > maxRGB {
		*colorValue = maxRGB
	}
}

func makeColor(red, green, blue, alpha int) rl.Color {
	return rl.NewColor(uint8(red), uint8(green), uint8(blue), uint8(alpha))
}

func drawDecaySettings(position rl.Vector2, decayValue *bool) (bool, rl.Vector2) {
	rg.Label(rl.NewRectangle(position.X, position.Y, 50, 20), "Decay")
	position.Y += 20
	*decayValue = rg.CheckBox(rl.NewRectangle(position.X, position.Y, 50, 20), *decayValue)
	return *decayValue, position
}

func drawSquares(squareContets map[GridCord]SquareInfo, fadeMode bool) {
	for cord, square := range squareContets {
		if decayMode {
			if timeLeft := time.Now().Sub(square.CreatedAt); timeLeft < decayTime {
				// scale for alpha
				var alpha = float32(1.0)
				if fadeMode {
					alpha = 1.0 - float32(timeLeft.Nanoseconds())/float32(decayTime.Nanoseconds())
				}
				rl.DrawRectangleV(square.Origin, rl.NewVector2(spacingFloat, spacingFloat), rl.Fade(square.Color, alpha))
			} else {
				delete(squareContets, cord)
			}
		} else {
			rl.DrawRectangleV(square.Origin, rl.NewVector2(spacingFloat, spacingFloat), square.Color)
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
		x >= gridWidth || y >= gridHeight {
		return SquareInfo{}, errors.New("Outside of screen bounds")
	}
	xPos := x / spacing
	yPos := y / spacing

	info := SquareInfo{
		GridCord:  GridCord{uint8(xPos), uint8(yPos)},
		Origin:    rl.NewVector2(float32(xPos*spacing), float32(yPos*spacing)),
		CreatedAt: time.Now(),
	}
	return info, nil
}
