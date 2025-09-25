
SOURCE CODE
-----------
tim@Timothys-MacBook-Air func_crash % cat crash_in_function.go 
1: package main

3: import "fmt"

5: // A small example that crashes inside a called function so you can practice
6: // stepping into/out and using `locals`, `print`, `bt`, `step`, `next`, `continue`.

8: // This file is deliberately a standalone example (different package main
9: // in a different folder) so it won't conflict with other files in the repo.

11: type Person struct {
12:	Name string
13: }

14: func main() {
15:	fmt.Println("program start")
16:	// Prepare some state you can inspect
17:	p := &Person{Name: "Alice"}
18:	_ = p // keep p to avoid unused variable error
19:	msg := prepare()
20:	fmt.Println("prepare returned:", msg)

22:	// Intentionally pass a nil pointer so the crash happens inside `process`.
23:	// This allows you to step into `callCrash` -> `process` and inspect locals.
24:	callCrash(nil)

26:	// Won't be reached
27:	fmt.Println("program end")
28: }

30: func prepare() string {
31:	n := 42
32:	s := "ready"
33:	fmt.Println("in prepare", n, s)
34:	return s
35: }

37: func callCrash(p *Person) {
38:	fmt.Println("in callCrash, p =", p)
39:	process(p)
40: }

42: func process(p *Person) {
43:	fmt.Println("in process, about to access p.Name")
44:	// Crash: nil pointer dereference when evaluating p.Name
45:	fmt.Println("person name:", p.Name)
46: }
tim@Timothys-MacBook-Air func_crash % 

COMPILE
-------
tim@Timothys-MacBook-Air func_crash % go build crash_in_function.go 
tim@Timothys-MacBook-Air func_crash % ls -l
total 4696
-rwxr-xr-x  1 tim  staff  2387970 Sep 25 11:45 crash_in_function
-rw-r--r--@ 1 tim  staff     1167 Sep 25 11:14 crash_in_function.go

DEBUG: Set Breakpoints
----------------------
tim@Timothys-MacBook-Air func_crash % dlv debug
Type 'help' for list of commands.

(dlv) break crash_in_function.go:16
Breakpoint 1 set at 0x1049a721c for main.main() ./crash_in_function.go:16

(dlv) break crash_in_function.go:19
Command failed: could not find statement at /Users/tim/Documents/workspace/go/w3schools/debug/func_crash/crash_in_function.go:19, please use a line with a statement

(dlv) break crash_in_function.go:21
Breakpoint 2 set at 0x1049a7294 for main.main() ./crash_in_function.go:21

DEBUG: Execute Program
----------------------
(dlv) continue
> [Breakpoint 1] main.main() ./crash_in_function.go:16 (hits goroutine(1):1 total:1) (PC: 0x100fe321c)
    11:	type Person struct {
    12:		Name string
    13:	}
    14:	
    15:	func main() {
=>  16:		fmt.Println("program start")
    17:		// Prepare some state you can inspect
    18:		p := &Person{Name: "Alice"}
    19:		_ = p // keep p to avoid unused variable error
    20:		msg := prepare()
    21:		fmt.Println("prepare returned:", msg)
(dlv) continue
program start
in prepare 42 ready
> [Breakpoint 2] main.main() ./crash_in_function.go:21 (hits goroutine(1):1 total:1) (PC: 0x100fe3294)
    16:		fmt.Println("program start")
    17:		// Prepare some state you can inspect
    18:		p := &Person{Name: "Alice"}
    19:		_ = p // keep p to avoid unused variable error
    20:		msg := prepare()
=>  21:		fmt.Println("prepare returned:", msg)
    22:	
    23:		// Intentionally pass a nil pointer so the crash happens inside `process`.
    24:		// This allows you to step into `callCrash` -> `process` and inspect locals.
    25:		callCrash(nil)
    26:	

DEBUG: Inspect Variables
------------------------
(dlv) print p
(*main.Person)(0x14000109ec8)
*main.Person {Name: "Alice"}
(dlv) print msg
"ready"
(dlv) locals
p = (*main.Person)(0x1400012dec8)
msg = "ready"

DEBUG: Crash Program
--------------------
(dlv) continue
prepare returned: ready
in callCrash, p = <nil>
in process, about to access p.Name
> [unrecovered-panic] runtime.fatalpanic() /opt/homebrew/Cellar/go/1.25.1/libexec/src/runtime/panic.go:1298 (hits goroutine(1):1 total:1) (PC: 0x100f61fb0)
Warning: debugging optimized function
	runtime.curg._panic.arg: interface {}(string) "runtime error: invalid memory address or nil pointer dereference"
  1293:	// fatalpanic implements an unrecoverable panic. It is like fatalthrow, except
  1294:	// that if msgs != nil, fatalpanic also prints panic messages and decrements
  1295:	// runningPanicDefers once main is blocked from exiting.
  1296:	//
  1297:	//go:nosplit
=>1298:	func fatalpanic(msgs *_panic) {
  1299:		pc := sys.GetCallerPC()
  1300:		sp := sys.GetCallerSP()
  1301:		gp := getg()
  1302:		var docrash bool
  1303:		// Switch to the system stack to avoid any stack growth, which
(dlv) bt
0  0x0000000100f61fb0 in runtime.fatalpanic
   at /opt/homebrew/Cellar/go/1.25.1/libexec/src/runtime/panic.go:1298
1  0x0000000100f91c78 in runtime.gopanic
   at /opt/homebrew/Cellar/go/1.25.1/libexec/src/runtime/panic.go:802
2  0x0000000100f60ae4 in runtime.panicmem
   at /opt/homebrew/Cellar/go/1.25.1/libexec/src/runtime/panic.go:262
3  0x0000000100f9353c in runtime.sigpanic
   at /opt/homebrew/Cellar/go/1.25.1/libexec/src/runtime/signal_unix.go:925
4  0x0000000100fe3610 in main.process
   at ./crash_in_function.go:46
5  0x000000010101f6c0 in ???
   at ?:-1
6  0x0000000100fe3338 in main.main
   at ./crash_in_function.go:25
(dlv) exit

