package bus

import (
	"errors"
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
  BusStateStop  BusState = 0
  BusStateRun   BusState = 1
)


//
// 运行中的总线实例对象
//
type Bus interface {
  start() error
  stop() error
  OnData(s *Slot, r DataRecv) error
  SendMsg(s *Slot, d DataWrap) error
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

  if _, has := busInstance[id]; has {
    return BusStateRun
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
// typeName -- 总线类型, 在 busInfos 中
//
func StartBus(id string, typeName string) error {
  busMutex.Lock()
  defer busMutex.Unlock()
  
  if _, has := busInstance[id]; has {
    return errors.New("总线已经启动 "+ id)
  }

  ct, has := bus_type_register[typeName]
  if !has {
    return errors.New("无效的总线类型 "+ typeName)
  }

  bus, err := ct.Create()
  if err != nil {
    return err
  }
  if err := bus.start(); err != nil {
    return err
  }
  busInstance[id] = bus
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