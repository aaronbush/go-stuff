// Copyright © 2019 Aaron S. Bush <asb.bush@gmail.com>
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
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

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
	fps           int32
	numRows       int32
	numColumns    int32
	spacing       int32
	spacingFloat  float32
	maxBrightness float32
	decayTime     time.Duration
	gridHeight    int32
	gridWidth     int32
	gridColor     = rl.RayWhite
	decayMode     = false
	binaryLog     string
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

	paintCmd.Flags().Int32VarP(&fps, "fps", "f", 30, "frames per second")
	paintCmd.Flags().Int32VarP(&numRows, "rows", "r", 40, "number of rows")
	paintCmd.Flags().Int32VarP(&numColumns, "columns", "c", 20, "number of columns")
	paintCmd.Flags().Int32VarP(&spacing, "spacing", "s", 20, "cell spacing")
	paintCmd.Flags().Float32VarP(&maxBrightness, "brightness", "b", 50, "max brightness")
	paintCmd.Flags().DurationVarP(&decayTime, "decayTime", "t", 3*time.Second, "decay time (seconds)")
	paintCmd.Flags().StringVarP(&binaryLog, "binaryLog", "l", "test.data", "binary log file name")

	log.SetLevel(log.DebugLevel)
}

func paint(cmd *cobra.Command, args []string) error {
	gridOrigin := rl.NewVector2(0, 3)
	gridHeight = numRows * spacing
	gridWidth = numColumns * spacing

	statusBarOrigin := rl.NewVector2(gridOrigin.X, gridOrigin.Y+float32(gridHeight))
	rightControlOrigin := rl.NewVector2(gridOrigin.X+float32(gridWidth), gridOrigin.Y)
	stausBarHeight := int32(60)
	rightControlWidth := int32(60)

	windowHeight := gridHeight + stausBarHeight
	windowWidth := gridWidth + rightControlWidth

	gridContents := makeGridContents(gridOrigin, uint8(numRows), uint8(numColumns))
	spacingFloat = float32(spacing)

	redValue, greenValue, blueValue := new(int), new(int), new(int)
	*redValue = 255

	fadeMode := false
	logMode := false
	floodFillMode := false

	// Open a new file for writing only
	binaryLogFile, err := os.OpenFile(
		binaryLog,
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0666,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer binaryLogFile.Close()

	rl.InitWindow(windowWidth, windowHeight, "pixel drawing")
	rg.LoadGuiStyle("cmd/styles/monokai.style")

	rl.SetTargetFPS(fps)

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.Blank)

		drawColor, decayOrigin := drawColorInputs(rightControlOrigin, redValue, greenValue, blueValue)
		decayMode, _ = drawDecaySettings(decayOrigin, &decayMode)

		drawSquares(gridContents, fadeMode, decayMode)

		drawGrid(gridOrigin, numRows, numColumns) // after colors are drawn to keep grid lines

		statusText := fmt.Sprintf("fade:%t, log:%t, decay:%t\nFPS: %.1f (%.03f)", fadeMode, logMode, decayMode, rl.GetFPS(), rl.GetFrameTime())
		rl.DrawText(statusText, int32(statusBarOrigin.X+3), int32(statusBarOrigin.Y), 12, rl.Gray)

		if logMode {
			exportSquares(binaryLogFile, gridContents, fadeMode, decayMode)
		}

		if rl.IsKeyPressed(rl.KeyF) {
			fadeMode = !fadeMode
		}

		if rl.IsKeyPressed(rl.KeyL) {
			logMode = !logMode
		}

		if rl.IsKeyPressed(rl.KeyB) {
			floodFillMode = !floodFillMode
		}

		if rl.IsKeyPressed(rl.KeyC) {
			for k, v := range gridContents {
				v.Color = rl.Blank
				gridContents[k] = v
			}
		}

		mousePos := rl.GetMousePosition()
		gridCord, err := gridCordFromMouseCord(gridOrigin, mousePos)

		if err == nil {
			squareInfo, ok := gridContents[gridCord]
			if ok {
				if rl.IsMouseButtonDown(rl.MouseRightButton) {
					squareInfo.Color = rl.Blank
				} else if rl.IsMouseButtonDown(rl.MouseLeftButton) {
					if floodFillMode {
						floodFill(gridContents, squareInfo, drawColor)
					}
					squareInfo.Color = drawColor // might be redundant if we just filled it
					squareInfo.CreatedAt = time.Now()
				}
				gridContents[squareInfo.GridCord] = squareInfo
			} else {
				log.Fatal("not found", gridCord)
			}
		}

		rl.EndDrawing()
	}

	rl.CloseWindow()
	return nil
}

func drawGrid(gridOrigin rl.Vector2, numRows, numColumns int32) {
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

func drawSquares(gridContents map[GridCord]SquareInfo, fadeMode, decayMode bool) {
	for cord, square := range gridContents {
		color := fadeAndDecay(square, fadeMode, decayMode)
		square.Color = color
		gridContents[cord] = square
		rl.DrawRectangleV(square.Origin, rl.NewVector2(spacingFloat, spacingFloat), color)
	}
}

func exportSquares(file *os.File, squares map[GridCord]SquareInfo, fadeMode, decayMode bool) {
	ledInfo := frame.LEDInfo{}
	var binBuf bytes.Buffer
	for _, square := range squares {
		color := fadeAndDecay(square, fadeMode, decayMode)

		ledInfo.Column = square.GridCord.Column
		ledInfo.Row = square.GridCord.Row
		ledInfo.Brightness = color.A
		ledInfo.Red = square.Color.R
		ledInfo.Blue = square.Color.B
		ledInfo.Green = square.Color.G

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

func fadeAndDecay(square SquareInfo, fadeMode, decayMode bool) rl.Color {
	if decayMode {
		if timeLeft := time.Now().Sub(square.CreatedAt); timeLeft < decayTime {
			// scale for alpha
			var alpha = float32(1.0)
			if fadeMode {
				alpha = 1.0 - float32(timeLeft.Nanoseconds())/float32(decayTime.Nanoseconds())
			}
			return rl.Fade(square.Color, alpha)
		}
		return rl.Blank
	}
	return square.Color
}

func makeGridContents(gridOrigin rl.Vector2, numRows, numColumns uint8) map[GridCord]SquareInfo {
	gridContents := make(map[GridCord]SquareInfo)

	for r := uint8(0); r < numRows; r++ {
		for c := uint8(0); c < numColumns; c++ {
			squareInfo := makeSquare(gridOrigin, r, c)
			gridContents[squareInfo.GridCord] = squareInfo
		}
	}
	return gridContents
}

func makeSquare(gridOrigin rl.Vector2, row, column uint8) SquareInfo {
	colInt32, rowInt32 := int32(column), int32(row)

	return SquareInfo{
		GridCord: GridCord{Row: row, Column: column},
		Origin:   rl.NewVector2(float32(colInt32*spacing)+gridOrigin.X, float32(rowInt32*spacing)+gridOrigin.Y),
	}
}

func gridCordFromMouseCord(gridOrigin, mouseVec rl.Vector2) (GridCord, error) {
	x := int32(mouseVec.X + gridOrigin.X)
	y := int32(mouseVec.Y + gridOrigin.Y)

	if x <= 0 || y <= 0 ||
		x >= gridWidth || y >= gridHeight {
		return GridCord{}, errors.New("Outside of screen bounds")
	}
	xPos := x / spacing
	yPos := y / spacing

	return GridCord{Row: uint8(yPos), Column: uint8(xPos)}, nil
}

/*
Flood-fill (node, target-color, replacement-color):
 1. If target-color is equal to replacement-color, return.
 2. If color of node is not equal to target-color, return.
 3. Set Q to the empty queue.
 4. Add node to Q.
 5. For each element N of Q:
 6.     Set w and e equal to N.
 7.     Move w to the west until the color of the node to the west of w no longer matches target-color.
 8.     Move e to the east until the color of the node to the east of e no longer matches target-color.
 9.     For each node n between w and e:
10.         Set the color of n to replacement-color.
11.         If the color of the node to the north of n is target-color, add that node to Q.
12.         If the color of the node to the south of n is target-color, add that node to Q.
13. Continue looping until Q is exhausted.
14. Return.
*/
func floodFill(gridContents map[GridCord]SquareInfo, square SquareInfo, newColor rl.Color) int {
	squaresChanged := 0
	targetColor := square.Color
	if targetColor == newColor {
		return squaresChanged
	}
	var queue []SquareInfo
	queue = append(queue, square)

	// for _, n := range queue {
	for i := 0; i < len(queue); i++ {

		west, east := queue[i], queue[i]

		// Go West
		west = furthestSquare(gridContents, west, targetColor, func(a, b uint8) uint8 { return a - b })
		east = furthestSquare(gridContents, east, targetColor, func(a, b uint8) uint8 { return a + b })

		// set nodes in between to newColor
		for wCol, eCol := west.GridCord.Column, east.GridCord.Column; wCol <= eCol; wCol++ {
			currentCord := GridCord{Row: west.GridCord.Row, Column: wCol}
			newSquare, ok := gridContents[currentCord]
			if !ok {
				panic(fmt.Sprintf("should have found square at %v", currentCord))
			}
			newSquare.Color = newColor
			gridContents[currentCord] = newSquare
			squaresChanged++

			// check to noth
			if currentCord.Row > 0 {
				northCord := currentCord
				northCord.Row--
				northSquare, ok := gridContents[northCord]
				if ok && northSquare.Color == targetColor {
					//	fmt.Printf("Adding %v to the north\n", northSquare)
					queue = append(queue, northSquare)
				}
			}
			// check to the south
			if currentCord.Row < uint8(numRows)-1 {
				southCord := currentCord
				southCord.Row++
				southSquare, ok := gridContents[southCord]
				if ok && southSquare.Color == targetColor {
					//fmt.Printf("Adding %v to the south\n", southSquare)
					queue = append(queue, southSquare)
				}
			}
		}
	}
	return squaresChanged
}

func furthestSquare(gridContents map[GridCord]SquareInfo, startingPoint SquareInfo, targetColor rl.Color, f func(a, b uint8) uint8) SquareInfo {
	result := startingPoint
	//fmt.Printf("Started at: %+v for target color: %+v\n", startingPoint, targetColor)
	for {
		squareNext := result.GridCord
		squareNext.Column = f(squareNext.Column, 1)
		//	fmt.Printf("square to squareNext: %+v\n", squareNext)
		tSquare, ok := gridContents[squareNext]
		if !ok || tSquare.Color != targetColor {
			//		fmt.Printf("stopped at %+v/%t\n", tSquare, ok)
			break
		}
		result = tSquare
		//	fmt.Printf("updated squareNext to: %+v\n", tSquare)
	}

	//	fmt.Printf("Ended at: %+v\n", result)
	return result
}
