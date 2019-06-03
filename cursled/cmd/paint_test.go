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
	"fmt"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func makeGridLine(numColumns uint8) map[GridCord]SquareInfo {
	result := make(map[GridCord]SquareInfo)
	for c := uint8(0); c < numColumns; c++ {
		cord := GridCord{Column: c, Row: 0}
		result[cord] = SquareInfo{Color: rl.Black, GridCord: cord}
	}
	fmt.Println(result)
	return result
}

func Test_floodFill(t *testing.T) {
	type args struct {
		gridContents map[GridCord]SquareInfo
		square       SquareInfo
		newColor     rl.Color
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO: Add test cases.  This is for debugging only right now.
		{"fills to west edge", args{
			gridContents: makeGridLine(3),
			square: SquareInfo{
				GridCord: GridCord{Row: 0, Column: 2},
				Color:    rl.Black},
			newColor: rl.Red},
			2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := floodFill(tt.args.gridContents, tt.args.square, tt.args.newColor); got != tt.want {
				t.Errorf("floodFill() = %v, want %v", got, tt.want)
			}
		})
	}
}
