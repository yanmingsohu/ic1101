package bus

import (
	"errors"
	"ic1101/src/core"
	"log"
	"net/url"
	"runtime"
	"sync"
	"time"
)


//
// 该接口是面向总线使用者的, 尽可能少的暴露方法和属性
//
type BusInfo struct {
  log *Log

  // 总线id
  id string
  uri *url.URL
  // 总线类型, 在 bus.busInfos 中
  typeName string
  // 定时抓取数据的定时器
  tk core.Tick
  // 消息接收器
  event BusEvent
  sync  *sync.Mutex

  st BusState
  bs Bus
  ps SlotParser

  // 数据插槽配置, {插槽: 数据名}
  datas []Slot
  ctrls []ctrl_slot
}


type ctrl_slot struct {
  slot Slot
  tk   core.Tick
  data DataWrap
}


//
// 创建一个用于启动总线的数据对象
//
// uri -- 用于创建服务器/连接客户端
// id  -- 总线id
// typ -- 总线类型
// tk  -- 总线数据定时器, 如果该定时器停止, 则所有相关任务都会停止, 总线退出
// ev  -- 事件接收器
//
func NewInfo(uri, id, typ string, tk core.Tick, ev BusEvent) (*BusInfo, error) {
  if id == "" {
    return nil, errors.New("id 不能为空")
  }
  sp, err := GetSlotParser(typ)
  if err != nil {
    return nil, err
  }
  if tk == nil {
    return nil, errors.New("必须提供定时器")
  }
  if tk.IsRunning() {
    return nil, errors.New("不能是已经启动的定时器")
  }
  if ev == nil {
    return nil, errors.New("必须提供事件监听器")
  }
  u, err := sp.ParseURI(uri)
  if err != nil {
    return nil, err
  }
  return &BusInfo{
    log       : NewLog(MaxLogCount, id),
    id        : id, 
    uri       : u, 
    typeName  : typ, 
    tk        : tk, 
    event     : ev, 
    st        : BusStateStartup, 
    bs        : nil, 
    ps        : sp, 
    datas     : make([]Slot, 0, 10), 
    ctrls     : make([]ctrl_slot, 0, 10), 
    sync      : new(sync.Mutex),
  } , nil
}


//
// 添加一个数据接口, 总线运行后从该接口取出数据发送到事件监听器
//
func (i *BusInfo) AddData(s Slot) error {
  if i.st != BusStateStartup {
    return errors.New("总线已经启动, 不能修改状态")
  }
  i.datas = append(i.datas, s)
  return nil
}


//
// 添加一个发送任务到总线接口, 当定时器 tk 开始执行, value 被发送到控制槽
//
func (i *BusInfo) AddCtrl(s Slot, tk core.Tick, value DataWrap) error {
  if i.st != BusStateStartup {
    return errors.New("总线已经启动, 不能修改状态")
  }
  if tk.IsRunning() {
    return errors.New("不能是已经启动的定时器")
  }
  i.ctrls = append(i.ctrls, ctrl_slot{s, tk, value})
  return nil
}


//
// 立即发送控制指令
//
func (i *BusInfo) SendCtrl(s Slot, value DataWrap, t *time.Time) error {
  i.sync.Lock()
  defer i.sync.Unlock()
  defer i.relife("总线发送控制")

  if i.st < BusStateStartup {
    return errors.New("总线没有运行, 不能发送控制指令")
  }
  err := i.bs.SendCtrl(s, value, t)
  if err == nil {
    i.event.OnCtrlSended(s, t)
  }
  return err
}


func (i *BusInfo) ParseSlot(s string) (Slot, error) {
  return i.ps.ParseSlot(s)
}


func (i *BusInfo) start(b Bus) error {
  if i.st != BusStateStartup {
    return errors.New("总线已经启动")
  }
  real := &bus_real_impl{i}
  if err := b.Start(real); err != nil {
    i.st = BusStateFailStart
    return err
  }
  i.bs = b

  for _, ctrl := range i.ctrls {
    i._ctrl_thread(ctrl)
  }

  i.st = BusStateSleep
  i.tk.Start(func() {
    i.sync.Lock()
    defer i.sync.Unlock()
    defer i.relife("总线同步数据")

    i.st = BusStateTask
    t := time.Now()
    b.SyncData(real, &t)
    i.st = BusStateSleep
  }, func() {
    i.stop()
  })
  return nil
}


func (i *BusInfo) stop() {
  if i.st <= BusStateStop {
    return
  }
  defer i.event.OnStopped()
  i.st = BusStateShutdown

  for _, ctrl := range i.ctrls {
    if ctrl.tk.IsRunning() {
      ctrl.tk.Stop()
    }
  }

  if i.tk.IsRunning() {
    i.tk.Stop()
  }
  if i.bs != nil {
    real := &bus_real_impl{i}
    i.bs.Stop(real)
  }
  i.st = BusStateStop
}


func (i *BusInfo) _ctrl_thread(c ctrl_slot) {
  c.tk.Start(func() {
    t := time.Now()
    if err := i.SendCtrl(c.slot, c.data, &t); err != nil {
      i.Log("发送控制失败", c.slot.String(), err)
    }
  }, func() {
    i.event.OnCtrlExit(c.slot)
  })
}


//
// 在 panic 中恢复, 在 defer 中调用
//
func (i *BusInfo) relife(action string) {
  if err := recover(); err != nil {
    i.Log(action, "发生了异常", err)

    var buf [4096]byte
    n := runtime.Stack(buf[:], false)
    log.Println("PANIC ==>", err, string(buf[:n]))
  }
}


func (i *BusInfo) State() BusState {
  return i.st
}


func (i *BusInfo) Log(msg ...interface{}) {
  i.log.Log(msg...)
}


func (i *BusInfo) GetLog() []string {
  return i.log.GetLog()
}