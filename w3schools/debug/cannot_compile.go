package main

import "fmt"

func main() {
	fmt.Println("This program will not compile")
	x := 42      // Valid declaration
	y := "hello" // Valid declaration
	z := x + y   // Error: cannot add int and string (mismatched types)
}
