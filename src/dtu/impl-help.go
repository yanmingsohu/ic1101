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
package dtu

import (
	"errors"
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
    return nil, errors.New("目标不存在")
  }
  if ctx.Closed() {
    delete(h.CtxMap, id)
    return nil, errors.New("目标已经关闭")
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