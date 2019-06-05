package main

import "fmt"

type I interface {
	m1()
	m2(int)
}

type T string

func main() {
	var t T = ""
	doM1(t)
	doM2(t)
}

func doM1(i I) {
	i.m1()
}

func doM2(i I) {
	i.m2(12)
}

func (t T) m1() {
	return
}

func (t T) m2(x int) {
	fmt.Println(x)
}
