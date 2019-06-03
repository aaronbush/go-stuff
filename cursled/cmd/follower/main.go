package main

import "fmt"

var f float64

type Foo struct {
	word string
}

func main() {
	var s []int

	s = append(s, 0)

	for i := 0; i < len(s); i++ {
		fmt.Println(s[i])
		s = append(s, i+1)
		//	i++
		// if i == len(s) {
		// break
		// }
	}
}
