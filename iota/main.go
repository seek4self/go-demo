package main

import (
	"fmt"
)

const (
	azero = iota
	aone  = iota
)

const (
	info  = "processing" // 0
	ss    = 4            // 1
	bzero = iota         // 2
	bone  = iota         // 3
)

func main() {
	fmt.Println(azero, aone) //prints: 0 1
	fmt.Println(bzero, bone) //prints: 2 3
}
