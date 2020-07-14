package dtu

import "errors"


//
// Context 复制类, 提供基本功能
//
type CtxHelp struct {
  bind   map[string]interface{}
  ID     int
  closed bool
  impl   Impl
}


func (c *CtxHelp) Id() int {
  return c.ID
}


func (c *CtxHelp) InitHelp(id int, di Impl) {
  c.ID = id
  c.bind = make(map[string]interface{})
  c.closed = false
  c.impl = di
}


func (c *CtxHelp) Get(name string) (interface{}, error) {
  if c.closed {
    return nil, errors.New("上下文已经关闭")
  }
  ret, has := c.bind[name]
  if !has {
    return nil, errors.New("变量不存在 "+ name)
  }
  return ret, nil
}


func (c *CtxHelp) Bind(name string, data interface{}) error {
  if c.closed {
    return errors.New("上下文已经关闭")
  }
  c.bind[name] = data
  return nil
}


//
// 通常在 Close() 中调用该方法, 关闭 help 的所有资源
// 如果关闭成功返回 true, 已经关闭返回 false
//
func (c *CtxHelp) CloseHelp() bool {
  if c.closed {
    return false
  }
  c.closed = true
  c.bind = nil
  c.impl.CloseContext(c.ID)
  return true
}