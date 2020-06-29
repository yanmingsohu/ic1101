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
  BusStateStop      BusState = 0 
  // 正在启动
  BusStateStartup   BusState = 1
  // 休眠中, 等待计时器
  BusStateSleep     BusState = 2
  // 执行任务
  BusStateTask      BusState = 3
)


//
// 运行中的总线实例对象
//
type Bus interface {
  start(*BusInfo) error
  stop() error
  OnData(s *Slot, r DataRecv) error
  SendMsg(s *Slot, d DataWrap) error
  State() BusState
}


type BusInfo struct {
  // 总线id
  Id string
  // 总线类型, 在 bus.busInfos 中
  TypeName string
  // 定时抓取数据的定时器
  Tk  *core.Tick
}


//
// 总线数据/控制槽
//
type Slot struct {
}


//
// 总线创建接口
//
type BusCreator interface {
  Create() (Bus, error)
}


//
// 数据包装器接口
//
type DataWrap interface {
  Int()     int
  Int64()   int64
  String()  string
  Float()   float32
}


//
// 数据接收器接口
//
type DataRecv interface {
  //
  // 接受总线发送的数据
  //
  OnData(t *time.Time, d *DataWrap)
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

  bus, err := ct.Create()
  if err != nil {
    return err
  }
  if err := bus.start(info); err != nil {
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
  if err := bus.stop(); err != nil {
    return err
  }
  delete(busInstance, id)
  return nil
}