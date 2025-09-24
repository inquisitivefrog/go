package main

import "fmt"

type Person struct {
    Name string
}

func main() {
    var p *Person // Declare a pointer to Person, but it's nil
    fmt.Println("Program started")
    // This will cause a runtime panic: nil pointer dereference
    fmt.Println(p.Name)
    fmt.Println("This line will not be reached")
}
