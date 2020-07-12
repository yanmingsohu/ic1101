package test

import (
	"ic1101/src/core"
	"testing"

	"github.com/dop251/goja"
)

const code = `
function aa(x) {
  return x;
}
function hello(x) {
  return new Date() + x;
}
`

func TestJS (t *testing.T) {
  sr := core.ScriptRuntime{}
  if err := sr.Compile("test.js", code); err != nil {
    t.Fatal(err)
  }
  if err := sr.InitVM(); err != nil {
    t.Fatal(err)
  }
  sr.VM().Set("tt", func (fc goja.FunctionCall) goja.Value {
    i := fc.Argument(0).ToInteger() + 10000
    return sr.VM().ToValue(i)
  })
  aa := sr.VM().Get("aa")
  t.Log(aa)

  hello, err := sr.GetFunc("hello")
  if err != nil {
    t.Fatal(err)
  }
  ret, err := hello(nil, sr.VM().ToValue(1))
  t.Log(ret, err)

  _, err = sr.GetFunc("noexist");
  if err == nil {
    t.Fatal()
  }
}