package bus

import (
	"fmt"
	"log"
	"sync"
	"time"
)


type Log struct {
  logs  []string
  logsy *sync.RWMutex
  id    string
}


func NewLog(max int, id string) *Log {
  l := Log{
    logs  : make([]string, 0, max),
    logsy : new(sync.RWMutex),
    id    : id,
  }
  return &l
}


func (i *Log) GetLog() []string {
  i.logsy.RLock()
  defer i.logsy.RUnlock()
  return i.logs[:]
}


// 插入新的日志, 删除超过 MaxLogCount 的部分
func (i *Log) Log(msg ...interface{}) {
  i.logsy.Lock()
  defer i.logsy.Unlock()

  s := fmt.Sprint(msg...)
  i.logs = append(i.logs, time.Now().Format(time.RFC3339) +" "+ s)
  if len(i.logs) > MaxLogCount {
    i.logs = i.logs[1:]
  }
  log.Println(i.id, s)
}


type NotRepeating struct {
  L Logger
  m map[uint64]bool
  s *sync.Mutex
}


//
// 记录日志, 并记录该日志的 id, 同样的 id 不会再打印
// 返回 true 说明状态已经改变 (打印了日志)
//
func (n *NotRepeating) Log(id uint64, msg ...interface{}) bool {
  n.s.Lock()
  defer n.s.Unlock()

  if no, has := n.m[id]; has == false || no == false {
    n.m[id] = true
    n.L.Log(msg...)
    return true
  }
  return false
}


//
// 恢复日志的记录, 并记录当前日志
// 返回 true 说明状态已经改变 (打印了日志)
//
func (n *NotRepeating) Recover(id uint64, msg ...interface{}) bool {
  n.s.Lock()
  defer n.s.Unlock()

  if no, has := n.m[id]; has == true && no == true {
    delete(n.m, id)
    n.L.Log(msg...)
    return true
  }
  return false
}


func NewNotRep(l Logger) NotRepeating {
  return NotRepeating{l, make(map[uint64]bool), new(sync.Mutex)}
}