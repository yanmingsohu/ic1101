package js

import "github.com/dop251/goja"


type JSValue interface {
  Value(i interface{}) goja.Value
  NewObject() *goja.Object
}


//
// 创建程序库的工厂
//
type LibFact interface {
  New(JSValue) interface{}
}