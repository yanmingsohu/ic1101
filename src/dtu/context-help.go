package dtu

import (
	"errors"
	"sync"
)


//
// Context 复制类, 提供基本功能
//
type CtxHelp struct {
  bind   map[string]interface{}
  ID     int
  closed bool
  sy     *sync.Mutex
}


func (c *CtxHelp) Id() int {
  return c.ID
}


func (c *CtxHelp) InitHelp(id int, di Impl) {
  c.ID = id
  c.bind = make(map[string]interface{})
  c.closed = false
  c.sy = new(sync.Mutex)
}


func (c *CtxHelp) Get(name string) (interface{}, error) {
  c.sy.Lock()
  defer c.sy.Unlock()

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
  c.sy.Lock()
  defer c.sy.Unlock()

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
  c.sy.Lock()
  defer c.sy.Unlock()
  
  if c.closed {
    return false
  }
  c.closed = true
  c.bind = nil
  return true
}


func (c *CtxHelp) Closed() bool {
  return c.closed
}