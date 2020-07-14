package js

import "github.com/dop251/goja"


type JSValue interface {
  // 把 go 对象包装为 js 对象
  Value(i interface{}) goja.Value
  // 创建一个空 js 对象
  NewObject() *goja.Object
}


//
// 创建程序库的工厂
//
type LibFact interface {
  // 创建程序库的实例
  New(JSValue) interface{}
}