package core

import "sync"

//
// 对某个对象的引用计数器
//
type RefCount struct {
  ref   map[string]int
  sy    *sync.RWMutex
}


func NewRefCount() *RefCount {
  return &RefCount{make(map[string]int), new(sync.RWMutex)}
}


//
// 返回对象的引用次数, 没有引用返回 0
//
func (r *RefCount) Count(s string) int {
  r.sy.RLock()
  defer r.sy.RUnlock()
  if c, has := r.ref[s]; has {
    return c
  }
  return 0
}


//
// 增加引用次数
//
func (r *RefCount) Add(s string) {
  r.sy.Lock()
  defer r.sy.Unlock()
  if _, has := r.ref[s]; has {
    r.ref[s]++
  } else {
    r.ref[s] = 1
  }
}


//
// 释放引用, 如果释放一个没有引用的对象则返回 false
//
func (r *RefCount) Free(s string) bool {
  r.sy.Lock()
  defer r.sy.Unlock()
  if c, has := r.ref[s]; has {
    if c > 1 {
      r.ref[s]--
    } else {
      delete(r.ref, s)
    }
    return true
  } else {
    return false
  }
}