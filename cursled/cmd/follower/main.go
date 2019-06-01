package main

import (
	"flag"
	"fmt"
)

var f float64

func main() {
	// f = float32(*flag.Int("f", 1, ""))
	flag.Float64Var(&f, "f", 1, "")
	f2 := flag.Float64("f2", 10, "")
	flag.Parse()
	fmt.Printf("%f\n", f)
	fmt.Printf("%f\n", *f2)
	fmt.Println(flag.Args())

	fmt.Println(int32(*myf()))
	fmt.Println(myf())
}

func myf() *int {
	x := 19
	return &x
}
