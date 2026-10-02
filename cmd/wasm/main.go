//go:build js && wasm

package main

import (
	"encoding/json"
	"syscall/js"

	"brotato-masher/internal/save"
)

func result(ok bool, value any) js.Value {
	payload, _ := json.Marshal(map[string]any{"ok": ok, "value": value})
	return js.ValueOf(string(payload))
}

func validate(this js.Value, args []js.Value) any {
	if len(args) != 1 {
		return result(false, "one JSON string is required")
	}
	_, err := save.Summary([]byte(args[0].String()))
	if err != nil {
		return result(false, err.Error())
	}
	return result(true, "valid Brotato v3 save")
}

func summary(this js.Value, args []js.Value) any {
	if len(args) != 1 {
		return result(false, "one JSON string is required")
	}
	value, err := save.Summary([]byte(args[0].String()))
	if err != nil {
		return result(false, err.Error())
	}
	return result(true, value)
}

func apply(this js.Value, args []js.Value) any {
	if len(args) != 2 {
		return result(false, "save JSON and edits JSON are required")
	}
	var edits map[string]any
	if err := json.Unmarshal([]byte(args[1].String()), &edits); err != nil {
		return result(false, "invalid edits JSON: "+err.Error())
	}
	output, err := save.Apply([]byte(args[0].String()), edits)
	if err != nil {
		return result(false, err.Error())
	}
	return result(true, string(output))
}

func main() {
	js.Global().Set("brotatoWasm", map[string]any{
		"validate": js.FuncOf(validate), "summary": js.FuncOf(summary), "apply": js.FuncOf(apply),
	})
	select {}
}
