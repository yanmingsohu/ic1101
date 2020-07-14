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