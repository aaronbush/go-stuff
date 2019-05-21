package lib

import (
	"errors"
	"math"
)

type Base int

type UnaryFunc func(float64) float64

func Make(name string, base Base) (UnaryFunc, error) {
	if base == 2 || base == 10 {
		return func(v float64) float64 {
			var result float64
			switch base {
			case 2:
				result = math.Log(v)
			case 10:
				result = math.Log10(v)
			}
			return result
		}, nil
	}
	return nil, errors.New("Unsupported Base")
}
