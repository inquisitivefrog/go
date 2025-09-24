package main

import "fmt"

type Person struct {
	Name string
}

func main() {
	var t *Person
	var p *Person // Declare a pointer to Person, but it's nil
	fmt.Println("Program started")
	t = &Person{Name: "Bruce Wayne"}
	fmt.Println(t.Name)
	// This will cause a runtime panic: nil pointer dereference
	fmt.Println(p.Name)
	fmt.Println("This line will not be reached")
}
