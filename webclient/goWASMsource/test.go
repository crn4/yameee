package main

import (
	"strings"
	"syscall/js"
)

func main() {
	js.Global().Set("testfunc", js.FuncOf(testGoFunc))
	<-make(chan bool)
}

func testGoFunc(this js.Value, inputs []js.Value) interface{} {
	message := inputs[0].String()
	return strings.ToLower(message)
}
