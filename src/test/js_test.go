/**
 *  Copyright 2023 Jing Yanming
 * 
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */
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