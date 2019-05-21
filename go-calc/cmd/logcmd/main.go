package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/aaronbush/go-stuff/go-calc/lib"
)

func main() {
	f, err := lib.Make("2", 2)
	if err != nil {
		panic("oops")
	}

	fl, err := strconv.ParseFloat(os.Args[1], 64)
	if err != nil {
		panic("did not supply valid float")
	}

	fmt.Println(f(fl))
}
