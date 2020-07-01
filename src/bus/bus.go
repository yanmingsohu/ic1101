package bus

import (
	"errors"
	"ic1101/src/core"
	"sync"
	"time"
)

//
// 运行中的总线实例
//
var busInstance = map[string]Bus{}
var busMutex = new(sync.RWMutex)

//
// 总线运行状态
//
type BusState int

const (
  // 总线没有启动
  BusStateStop      BusState = iota 
  // 正在启动
  BusStateStartup   BusState = iota 
  // 休眠中, 等待计时器
  BusStateSleep     BusState = iota 
  // 执行任务
  BusStateTask      BusState = iota 
)

//
// 槽的类型
//
type SlotType int

const (
  // 无效值
  SlotInvaild SlotType = iota
  // 数据槽
  SlotData    SlotType = iota
  // 控制槽
  SlotCtrl    SlotType = iota
)

//
// 运行中的总线实例对象
//
type Bus interface {
  // 启动总线, 失败返回 error
  start() error

  // 停止总线, 该方法返回后, 总线一定是停止的
  stop()

  // 发送控制指令, 发送失败返回 error
  SendCtrl(s Slot, d DataWrap) error

  // 返回总线状态
  State() BusState
}


type BusInfo struct {
  // 总线id
  Id string
  // 总线类型, 在 bus.busInfos 中
  TypeName string
  // 定时抓取数据的定时器
  Tk core.Tick
  // 数据接收器
  Recv DataRecv
  // 数据插槽配置, {插槽: 数据名}
  SlotConf []Slot
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
// 用于解析 slot, 数据槽格式包含 '数据/控制' 的描述
//
type SlotParser interface {
  // 通过字符串 (序列化的slot) 解析 slot 实例, 失败返回 error
  // 每种总线都有自己的 slot 格式
  ParseSlot(s string) (Slot, error)

  // 返回可读的对端口的描述字符串, 格式无效返回 error
  SlotDesc(s string) (string, error)
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
// 数据接收器接口
//
type DataRecv interface {
  //
  // 接受总线发送的数据
  //
  OnData(slot Slot, time *time.Time, data DataWrap)
}


//
// 返回总线状态
//
func GetBusState(id string) BusState {
  busMutex.RLock()
  defer busMutex.RUnlock()

  if bus, has := busInstance[id]; has {
    return bus.State()
  }
  return BusStateStop
}


//
// 返回运行中的总线
//
func GetBus(id string) (Bus, error) {
  busMutex.RLock()
  defer busMutex.RUnlock()

  if bus, has := busInstance[id]; has {
    return bus, nil
  }
  return nil, errors.New(id +" 引用的总线不存在")
}


//
// 启动总线, 成功返回 nil, 否则返回错误
// id       -- 总线id
// typeName -- 
//
func StartBus(info *BusInfo) error {
  busMutex.Lock()
  defer busMutex.Unlock()
  
  if _, has := busInstance[info.Id]; has {
    return errors.New("总线已经启动 "+ info.Id)
  }

  ct, has := bus_type_register[info.TypeName]
  if !has {
    return errors.New("无效的总线类型 "+ info.TypeName)
  }

  bus, err := ct.Create(info)
  if err != nil {
    return err
  }
  if err := bus.start(); err != nil {
    return err
  }
  busInstance[info.Id] = bus
  return nil
}


//
// 停止总线, 成功返回 nil, 否则返回 error
//
func StopBus(id string) error {
  busMutex.Lock()
  defer busMutex.Unlock()

  bus, has := busInstance[id]
  if !has {
    return errors.New("总线没有启动 "+ id)
  }
  bus.stop()
  delete(busInstance, id)
  return nil
}


//
// 返回对应总线类型的数据槽解析器
//
func GetSlotParser(typeName string) (SlotParser, error) {
  busMutex.RLock()
  defer busMutex.RUnlock()

  ct, has := bus_type_register[typeName]
  if !has {
    return nil, errors.New("不存在的总线类型 "+ typeName)
  }
  return ct, nil
}