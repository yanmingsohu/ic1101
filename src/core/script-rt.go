package core

import (
	"errors"
	"reflect"

	"github.com/dop251/goja"
)


//
// 已经编译的脚本, 线程不安全, 可以创建一个全空的对象.
//
type ScriptRuntime struct {
  vm        *goja.Runtime
  pj        *goja.Program
  ext       *goja.Object
  on_data   goja.Callable
}


//
// 编译脚本
//
func (s *ScriptRuntime) Compile(name, code string) error {
  proj, err := goja.Compile(name, "("+ code +")", true)
  if err != nil {
    return err
  }
  s.pj = proj
  return nil
}


//
// 初始化脚本框架
//
func (s *ScriptRuntime) InitObject() error {
  if s.pj == nil {
    return errors.New("没有程序被编译")
  }
  if s.vm == nil {
    vm := goja.New()
    res, err := vm.RunProgram(s.pj)
    if err != nil {
      return err
    }
    if res.ExportType().Kind() != reflect.Map {
      return errors.New("脚本没有导出对象")
    }
    s.ext = res.ToObject(vm)
    s.vm = vm
  }
  return nil
}


func (s *ScriptRuntime) GetFunc(name string) (goja.Callable, error) {
  v := s.ext.Get(name)
  af, is := goja.AssertFunction(v)
  if !is {
    return nil, errors.New(name +" 不是函数")
  }
  return af, nil
}


func (s *ScriptRuntime) This() *goja.Object {
  return s.ext
}


func (s *ScriptRuntime) VM() *goja.Runtime {
  return s.vm
}