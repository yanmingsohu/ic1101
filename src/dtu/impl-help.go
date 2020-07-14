package dtu

import (
	"errors"
	"strconv"
	"sync"
)


//
// Impl 接口辅助类, 提供基本功能, 该对象本身是一把锁
//
type ImplHelp struct {
  sy      *sync.Mutex
  CtxMap  map[int]Context
  Run     bool
}


func NewImplHelp() ImplHelp {
  return ImplHelp{
    sy      : new(sync.Mutex),
    CtxMap  : make(map[int]Context),
    Run     : true,
  }
}


func (h *ImplHelp) SaveContext(c Context) error {
  h.sy.Lock()
  defer h.sy.Unlock()
  
  if !h.Run {
    return errors.New("DTU 已经关闭")
  }

  old, has := h.CtxMap[c.Id()]
  if has {
    old.Close()
  }
  h.CtxMap[c.Id()] = c
  return nil
}


func (h *ImplHelp) GetContext(id int) (Context, error) {
  h.sy.Lock()
  defer h.sy.Unlock()

  if !h.Run {
    return nil, errors.New("DTU 已经关闭")
  }
  ctx, has := h.CtxMap[id]
  if !has {
    return nil, errors.New("上下文不存在"+ strconv.Itoa(id))
  }
  return ctx, nil
}


func (h *ImplHelp) CloseContext(id int) error {
  h.sy.Lock()
  defer h.sy.Unlock()

  if !h.Run {
    return errors.New("DTU 已经关闭")
  }
  delete(h.CtxMap, id)
  return nil
}