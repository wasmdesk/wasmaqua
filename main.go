// SPDX-License-Identifier: BSD-3-Clause
//
// Command wasmaqua is the in-browser compositor: it bakes the Ruby window
// manager (compositor.rb) into the WebAssembly binary with //go:embed and runs
// it on the embedded go-embedded-ruby interpreter. There is no separate .rb
// fetched at runtime — the program ships inside the wasm.
//
// wasmaqua is the macOS-Aqua-themed sibling of wasmbox: same external-client
// protocol (so the same hello/dock/quake/terminal/files wasm clients work
// unmodified), different window decoration (three traffic-light buttons on the
// LEFT, gradient titlebar).
//
//go:build js && wasm

package main

import (
	_ "embed"
	"os"
	"syscall/js"

	ruby "github.com/go-embedded-ruby/ruby"
)

//go:embed compositor.rb
var compositor string

func main() {
	// Run the embedded compositor. It installs DOM/Canvas event handlers and an
	// animation loop through the interpreter's JS bridge, which keep the VM (and,
	// via select{} below, the Go runtime) alive after Run returns.
	if err := ruby.Run(compositor, os.Stdout); err != nil {
		js.Global().Set("wasmaquaError", err.Error())
		return
	}
	js.Global().Set("wasmaquaReady", true)
	select {} // keep the runtime alive for the browser event/animation callbacks
}
