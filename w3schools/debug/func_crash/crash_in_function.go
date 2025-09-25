package main

import "fmt"

// A small example that crashes inside a called function so you can practice
// stepping into/out and using `locals`, `print`, `bt`, `step`, `next`, `continue`.

// This file is deliberately a standalone example (different package main
// in a different folder) so it won't conflict with other files in the repo.

type Person struct {
	Name string
}

func main() {
	fmt.Println("program start")
	// Prepare some state you can inspect
	p := &Person{Name: "Alice"}
	_ = p // keep p to avoid unused variable error
	msg := prepare()
	fmt.Println("prepare returned:", msg)

	// Intentionally pass a nil pointer so the crash happens inside `process`.
	// This allows you to step into `callCrash` -> `process` and inspect locals.
	callCrash(nil)

	// Won't be reached
	fmt.Println("program end")
}

func prepare() string {
	n := 42
	s := "ready"
	fmt.Println("in prepare", n, s)
	return s
}

func callCrash(p *Person) {
	fmt.Println("in callCrash, p =", p)
	process(p)
}

func process(p *Person) {
	fmt.Println("in process, about to access p.Name")
	// Crash: nil pointer dereference when evaluating p.Name
	fmt.Println("person name:", p.Name)
}
