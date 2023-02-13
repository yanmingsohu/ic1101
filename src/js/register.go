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
package js

import (
	"errors"
	"log"
)


var _js_register = make(map[string]LibFact)


//
// 注册一个类工厂, 任何冲突都会引发 panic
//
func Reg(name string, f LibFact) {
  if _, has := _js_register[name]; has {
    panic(errors.New("类库名称冲突 "+ name))
  }
  _js_register[name] = f
  log.Println("REG javascript lib", name)
}


func All(sr *ScriptRuntime, f func(string, interface{})) {
  for name, fact := range _js_register {
    lib := fact.New(sr)
    f(name, lib)
  }
}