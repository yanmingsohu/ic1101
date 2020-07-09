package core

import "sync"

// #cgo CFLAGS: -I${SRCDIR}/native
// #cgo LDFLAGS: ${SRCDIR}/build/native.a
// #include <main.h>
import "C"

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


//
// 变成一个单行字符串 (删掉回车/换行)
//
func Singleline(s string) string {
  buf := make([]rune, 0, 100)
  for _, ch := range s {
    if ch != '\n' && ch != '\r' {
      buf = append(buf, ch)
    }
  }
  return string(buf)
}


//
// 变成多行字符串
//
func Multiline(s string) string {
  column := 50
  buf := make([]rune, 0, len(s))
  i := 0

  for _, ch := range s {
    buf = append(buf, ch)
    i++
    if i > column {
      buf = append(buf, '\n')
      i = 0
    }
  }
  return string(buf)
}


//
// 读取公钥
//
func pick_session_info() []byte {
  return []byte(_cpu_core_info)
}


//
// 获取硬件加密信息
//
func pick_ref_count_by_user(string) []byte {
  return nil
}