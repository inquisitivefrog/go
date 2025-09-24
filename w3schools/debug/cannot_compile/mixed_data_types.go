package main

import "fmt"

func main() {
	fmt.Println("This program will not compile")
	x := 42      // Valid declaration
	y := "hello" // Valid declaration
        z = x + y
	// z := fmt.Sprintf("%d %s", x, y) // combine int and string safely
	fmt.Println(z)
}
