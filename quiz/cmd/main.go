package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

const (
	file string = "problems.csv"
)

func main() {
	file, err := os.Open(file)
	if err != nil {
		panic(err)
	}

	in := bufio.NewScanner(os.Stdin)

	var numCorrect = 0.0
	var numAsked = 0.0

	csv := csv.NewReader(file)

	for {
		rec, err := csv.Read()
		if err == io.EOF {
			break
		}
		fmt.Println(rec[0])
		fmt.Println("? ")
		numAsked++
		in.Scan()
		fmt.Printf("You said: %s\n", in.Text())
		if in.Text() == rec[1] {
			numCorrect++
		}
	}
	fmt.Printf("You got %.0f out of %.0f correct; score is %.1f%%\n", numCorrect, numAsked, (numCorrect/numAsked)*100)
}
