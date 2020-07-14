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

