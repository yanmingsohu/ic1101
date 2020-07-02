package bus

import (
	"errors"
	"fmt"
	"ic1101/src/core"
	"net/url"
	"sync"
	"time"
)

//
// 运行中的总线实例
//
var busInstance = map[string]*BusInfo{}
var busMutex = new(sync.RWMutex)
const MaxLogCount = 20

//
// 总线运行状态
//
type BusState int

const (
  // 总线没有启动
  BusStateStop      BusState = iota 
  BusStateStartup   BusState = iota 
  BusStateSleep     BusState = iota 
  BusStateTask      BusState = iota 
  BusStateShutdown  BusState = -1 
  BusStateFailStart BusState = -2
)

func (s BusState) String() string {
  switch s {
  case BusStateStop:
    return "停止"
  case BusStateStartup:
    return "正在启动"
  case BusStateSleep:
    return "休眠中, 等待计时器"
  case BusStateTask:
    return "执行任务"
  case BusStateShutdown:
    return "关闭中"
  case BusStateFailStart:
    return "启动失败"
  }
  return "无效"
}

//
// 槽的类型
//
type SlotType int

const (
  SlotInvaild SlotType = iota
  SlotData    SlotType = iota
  SlotCtrl    SlotType = iota
)

func (t SlotType) String() string {
  switch (t) {
  case SlotData:
    return "数据槽"
  case SlotCtrl:
    return "控制槽"
  }
  return "无效"
}


//
// 总线创建接口
//
type BusCreator interface {
  SlotParser
  // 创建总线实例
  Create(*BusInfo) (Bus, error)
  // 总线的显示名称
  Name() string
}


//
// 运行中的总线实例对象
//
type Bus interface {
  // 启动总线, 用于初始化数据, 失败返回 error
  start(i *BusInfo) error
  // 传送一次数据
  sync_data(*BusInfo, *time.Time) error
  // 停止总线, 该方法返回后, 总线一定是停止的
  stop(i *BusInfo)
  // 发送控制指令, 发送失败返回 error
  send_ctrl(s Slot, d DataWrap, t *time.Time) error
}


type BusInfo struct {
  // 总线id
  id string
  // 总线类型, 在 bus.busInfos 中
  typeName string
  // 定时抓取数据的定时器
  tk core.Tick
  // 消息接收器
  event BusEvent

  st BusState
  bs Bus

  // 数据插槽配置, {插槽: 数据名}
  datas []Slot
  ctrls []ctrl_slot
  logs  []string
}


type ctrl_slot struct {
  slot Slot
  tk   core.Tick
  data DataWrap
}


//
// 创建一个用于启动总线的数据对象
//
// id  -- 总线id
// typ -- 总线类型
// tk  -- 总线数据定时器, 如果该定时器停止, 则所有相关任务都会停止, 总线退出
// ev  -- 事件接收器
//
func NewInfo(id string, typ string, tk core.Tick, ev BusEvent) (*BusInfo, error) {
  if id == "" {
    return nil, errors.New("id 不能为空")
  }
  if !HasTypeName(typ) {
    return nil, errors.New("无效的总线类型 "+ typ)
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
  return &BusInfo{id, typ, tk, ev, BusStateStartup, 
      nil, make([]Slot, 0, 10), make([]ctrl_slot, 0, 10), 
      make([]string, 0, MaxLogCount)} , nil
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


func (i *BusInfo) start(b Bus) error {
  if i.st != BusStateStartup {
    return errors.New("总线已经启动")
  }
  if err := b.start(i); err != nil {
    i.st = BusStateStop
    return err
  }
  i.bs = b

  for _, ctrl := range i.ctrls {
    i._ctrl_thread(ctrl)
  }

  i.st = BusStateSleep
  i.tk.Start(func() {
    i.st = BusStateTask
    t := time.Now()
    b.sync_data(i, &t)
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
    i.bs.stop(i)
  }
  i.st = BusStateStop
}


func (i *BusInfo) _ctrl_thread(c ctrl_slot) {
  c.tk.Start(func() {
    t := time.Now()
    i.bs.send_ctrl(c.slot, c.data, &t)
    i.event.OnCtrlSended(c.slot, &t)
  }, func() {
    i.event.OnCtrlExit(c.slot)
  })
}


func (i *BusInfo) State() BusState {
  return i.st
}


func (i *BusInfo) GetLog() []string {
  return i.logs
}


// 插入新的日志, 删除超过 MaxLogCount 的部分
func (i *BusInfo) log(s string) {
  s = fmt.Sprintln(time.Now().Format(time.RFC3339), s)
  i.logs = append(i.logs, s)
  if len(i.logs) > MaxLogCount {
    i.logs = i.logs[1:]
  }
}


//
// 总线数据/控制槽
//
type Slot interface {
  // 返回插槽的唯一标识名称, 该标识是对 slot 描述的完整序列化
  // 与 BusCreator 中的 ParseSlot 对应.
  // 总线实例需要判断 '数据/控制' 类型
  String() string
  // 返回对插槽的可读文本
  Desc() string
  Type() SlotType
}


//
// 用于解析 slot, 数据槽格式包含 '数据/控制' 的描述
//
type SlotParser interface {
  // 通过字符串 (序列化的slot) 解析 slot 实例, 失败返回 error
  // 每种总线都有自己的 slot 格式
  ParseSlot(s string) (Slot, error)
  // 返回可读的对端口的描述字符串, 格式无效返回 error
  SlotDesc(s string) (string, error)
  // 解析 uri 用于创建服务器, 如果 uri 无效, 或不受支持返回 error
  ParseURI(uri string) (*url.URL, error)
}


//
// 数据包装器接口
//
type DataWrap interface {
  Int()     int
  Int64()   int64
  String()  string
  Float()   float32
  Bool()    bool
}


//
// 消息接收器接口
//
type BusEvent interface {
  // 接受总线发送的数据, [该方法由总线调用]
  OnData(slot Slot, time *time.Time, data DataWrap)
  // 总线上所有任务都终止了, 该方法被调用
  OnStopped()
  // 一个控制命令已经发出后该方法被调用
  OnCtrlSended(s Slot, t *time.Time)
  // 一个控制槽已经终止任务该方法被调用
  OnCtrlExit(s Slot)
}


//
// 返回总线状态
//
func GetBusState(id string) BusState {
  busMutex.RLock()
  defer busMutex.RUnlock()

  if info, has := busInstance[id]; has {
    return info.State()
  }
  return BusStateStop
}


//
// 返回运行中的总线
//
func GetBus(id string) (*BusInfo, error) {
  busMutex.RLock()
  defer busMutex.RUnlock()

  if info, has := busInstance[id]; has {
    return info, nil
  }
  return nil, errors.New(id +" 引用的总线不存在")
}


//
// 启动总线, 成功返回 nil, 否则返回错误
//
func StartBus(info *BusInfo) (error) {
  busMutex.Lock()
  defer busMutex.Unlock()
  
  if _, has := busInstance[info.id]; has {
    return errors.New("总线已经启动 "+ info.id)
  }

  ct, has := bus_type_register[info.typeName]
  if !has {
    return errors.New("无效的总线类型 "+ info.typeName)
  }

  bus, err := ct.Create(info)
  if err != nil {
    return err
  }
  if err := info.start(bus); err != nil {
    return err
  }
  busInstance[info.id] = info
  return nil
}


//
// 停止总线, 成功返回 nil, 否则返回 error
//
func StopBus(id string) error {
  busMutex.Lock()
  defer busMutex.Unlock()

  info, has := busInstance[id]
  if !has {
    return errors.New("总线没有启动 "+ id)
  }
  info.stop()
  delete(busInstance, id)
  return nil
}