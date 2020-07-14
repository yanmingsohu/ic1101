package core

import (
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/dop251/goja"
)

const (
  MB = 1024*1024
)


type JSValue interface {
  Value(i interface{}) goja.Value
  NewObject() *goja.Object
}


//
// 已经编译的脚本, 线程不安全, 可以创建一个全空的对象.
//
type ScriptRuntime struct {
  vm        *goja.Runtime
  pj        *goja.Program
  on_data   goja.Callable
}


//
// 编译脚本
//
func (s *ScriptRuntime) Compile(name, code string) error {
  proj, err := goja.Compile(name, code, true)
  if err != nil {
    return err
  }
  s.pj = proj
  return nil
}


//
// 初始化脚本框架
//
func (s *ScriptRuntime) InitVM() error {
  if s.pj == nil {
    return errors.New("没有程序被编译")
  }
  if s.vm == nil {
    vm := goja.New()
    _, err := vm.RunProgram(s.pj)
    if err != nil {
      return err
    }
    s.vm = vm
    s.installGlobalFunc()
  }
  return nil
}


//
// 找不到函数会返回错误
//
func (s *ScriptRuntime) GetFunc(name string) (goja.Callable, error) {
  v := s.vm.Get(name)
  af, is := goja.AssertFunction(v)
  if !is {
    return nil, errors.New("脚本没有定义 "+ name +" 函数")
  }
  return af, nil
}


//
// 返回脚本导出的对象
//
func (s *ScriptRuntime) This() *goja.Object {
  return s.vm.GlobalObject()
}


//
// 返回虚拟机
//
func (s *ScriptRuntime) VM() *goja.Runtime {
  return s.vm
}


//
// 对 goja.Runtime.ToValue() 的包装
//
func (s *ScriptRuntime) Value(i interface{}) goja.Value {
  return s.vm.ToValue(i)
}


func (s *ScriptRuntime) NewObject() *goja.Object {
  return s.vm.NewObject()
}

//
// 终止虚拟机中任务
//
func (s *ScriptRuntime) Stop(cause string) {
  s.vm.Interrupt(cause)
}


func (s *ScriptRuntime) installGlobalFunc() {
  s.vm.Set("http", &JSHttp{s})
}


type JSHttp struct {
  JSValue
}


func (h *JSHttp) Send(f goja.FunctionCall) goja.Value {
  url := f.Argument(0).String()
  if url == "" {
    panic(errors.New("URL 参数不能为空"))
  }
  go (func() {
    _, err := http.Get(url)
    if err != nil {
      log.Println("http.send", err)
      return
    }
  })()
  return h.Value(nil)
}


func (h *JSHttp) Get(f goja.FunctionCall) goja.Value {
  url := f.Argument(0).String()
  if url == "" {
    panic(errors.New("URL 参数不能为空"))
  }
  res, err := http.Get(url)
  if err != nil {
    panic(err)
  }
  len := res.ContentLength
  if len > 3*MB {
    len = 3*MB
  }

  ret := h.NewObject()
  buf := make([]byte, len)
  io.ReadFull(res.Body, buf)

  ret.Set("status", res.Status)
  ret.Set("body",   buf)
  ret.Set("header", res.Header)
  return ret
}


func (h *JSHttp) Post(f goja.FunctionCall) goja.Value {
  return h.Value(nil)
}