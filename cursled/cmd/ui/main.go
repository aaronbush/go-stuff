package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log"
	"os"
	"time"

	"github.com/aaronbush/go-stuff/cursled/frame"

	rl "github.com/gen2brain/raylib-go/raylib"
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

var maxBrightness = float32(50)
var gridRows = int32(40)
var gridColumns = int32(20)
var gridSpacing = int32(20)
var gridSpacingFloat = float32(gridSpacing)
var windowHeight = gridRows * gridSpacing
var windowWidth = gridColumns * gridSpacing
var decayTime = 3 * time.Second

func main() {
	windowHeight := gridRows * gridSpacing
	windowWidth := gridColumns * gridSpacing
	traceContents := make(map[GridCord]SquareInfo)
	drawnContents := make(map[GridCord]SquareInfo)
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
}

func drawGrid() {
	for r := int32(0); r <= windowHeight; r += gridSpacing {
		rl.DrawLine(0, r, windowWidth, r, rl.Black)
	}
	for c := int32(0); c <= windowWidth; c += gridSpacing {
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
			rl.DrawRectangleV(square.Vector2, rl.NewVector2(gridSpacingFloat, gridSpacingFloat), rl.Fade(square.Color, alpha))
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
	xPos := x / gridSpacing
	yPos := y / gridSpacing

	info := SquareInfo{
		GridCord:  GridCord{uint8(xPos), uint8(yPos)},
		Vector2:   rl.NewVector2(float32(xPos*gridSpacing), float32(yPos*gridSpacing)),
		CreatedAt: time.Now(),
	}
	return info, nil
}
