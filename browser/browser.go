// Package browser the JS browser implementations
package browser

import (
	"encoding/json"
	"reflect"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/grafana/sobek"
	"github.com/shiroyk/ski/js"
)

func init() {
	js.Register("browser", new(Browser))
}

type Browser struct{}

func (Browser) Instantiate(rt *sobek.Runtime) (sobek.Value, error) {
	return rt.ToValue(func(call sobek.ConstructorCall) *sobek.Object {
		return rt.ToValue(&browser{rod.New().ControlURL(call.Argument(0).String()).MustConnect()}).ToObject(rt)
	}), nil
}

// Browser module represents the browser. It doesn't depends on file system,
// it should work with remote browser seamlessly.
type browser struct { //nolint:var-naming
	*rod.Browser
}

// Page returns a new page
func (b *browser) Page(call sobek.FunctionCall, rt *sobek.Runtime) sobek.Value {
	if call.Argument(0).ExportType().Kind() == reflect.String {
		page := b.MustPage(call.Argument(0).String())
		return NewPage(page, rt)
	}

	target := toGoStruct[proto.TargetCreateTarget](call.Argument(0), rt)
	page, err := b.Browser.Page(target)
	if err != nil {
		js.Throw(rt, err)
	}
	return NewPage(page, rt)
}

// toGoStruct mapping the js object to golang struct.
func toGoStruct[T any](value sobek.Value, vm *sobek.Runtime) (t T) {
	if sobek.IsUndefined(value) {
		return
	}
	bytes, err := value.ToObject(vm).MarshalJSON()
	if err != nil {
		js.Throw(vm, err)
	}
	err = json.Unmarshal(bytes, &t)
	if err != nil {
		js.Throw(vm, err)
	}
	return
}

// toJSObject mapping the golang struct to js object.
func toJSObject(value any, vm *sobek.Runtime) sobek.Value {
	if value == nil {
		return sobek.Undefined()
	}
	bytes, err := json.Marshal(value)
	if err != nil {
		js.Throw(vm, err)
	}
	var obj map[string]any
	err = json.Unmarshal(bytes, &obj)
	if err != nil {
		js.Throw(vm, err)
	}
	return vm.ToValue(obj)
}
