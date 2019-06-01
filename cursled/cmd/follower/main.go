package main

import "fmt"

var f float64

type Foo struct {
	word string
}

func main() {
	m := make(map[int]Foo)
	m[1] = Foo{"a"}

	fmt.Println(m)

	f := m[1]
	f.word = "b"
	m[1] = f

	fmt.Println(m)
}
