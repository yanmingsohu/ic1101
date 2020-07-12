package bus_random

import (
	"errors"
	"fmt"
	"ic1101/src/bus"
	"math/rand"
	"net/url"
	"strconv"
	"time"
)

const RealRegNum = 5;


func init() {
  bus.InstallBus("random", &bus_random_ct{})
}


func _parse_random_slot(s string) (*bus_random_sl, error) {
  var t int
  var port int
  var tp bus.SlotType

  n, err := fmt.Sscanf(s, "%c#%d", &t, &port)
  if err != nil {
    return nil, err
  }
  if n != 2 {
    return nil, errors.New("无效的slot格式")
  }
  switch (t) {
  case 'D':
    tp = bus.SlotData
  case 'C':
    tp = bus.SlotCtrl
  default:
    return nil, errors.New("无效的类型字符")
  }
  return &bus_random_sl{port, tp}, nil
}


type bus_random_ct struct {
}


func (*bus_random_ct) Name() string {
  return "随机数(测试)"
}


func (*bus_random_ct) Create(i bus.BusReal) (bus.Bus, error) {
  rb := random_bus{ make([]bus.DataWrap, RealRegNum) }
  for i := 0; i < RealRegNum; i++ {
    rb.reg[i] = &bus.IntData{D:0}
  }
  return &rb, nil
}


// 接受任何字符串作为 slot
func (*bus_random_ct) ParseSlot(s string) (bus.Slot, error) {
  return _parse_random_slot(s)
}


func (*bus_random_ct) SlotDesc(s string) (string, error) {
  slot, err := _parse_random_slot(s)
  if err != nil {
    return "", err
  }
  return slot.Desc(), nil
}


func (*bus_random_ct) ParseURI(uri string) (*url.URL, error) {
  return &url.URL{}, nil
}


type bus_random_sl struct {
  port int
  tp bus.SlotType
}


func (s *bus_random_sl) String() string {
  if s.tp == bus.SlotData {
    return "D#"+ strconv.Itoa(s.port)
  } else {
    return "C#"+ strconv.Itoa(s.port)
  }
}


func (s *bus_random_sl) Desc() string {
  if s.port < RealRegNum {
    if s.tp == bus.SlotData {
      return "虚拟寄存器 "+ strconv.Itoa(s.port)
    } else {
      return "虚拟控制 "+ strconv.Itoa(s.port)
    }
  } else {
    if s.tp == bus.SlotData {
      return "随机数据 "+ strconv.Itoa(s.port)
    } else {
      return "空控制 "+ strconv.Itoa(s.port)
    }
  }
}


func (s *bus_random_sl) Type() bus.SlotType {
  return s.tp
}


type random_bus struct {
  reg []bus.DataWrap
}


func (r *random_bus) Start(i bus.BusReal) error {
  i.Log("总线启动")
  return nil
}


func (r *random_bus) SyncData(i bus.BusReal, t *time.Time) error {
  for _, s := range i.Datas() {
    slot := s.(*bus_random_sl)
    if slot.port < RealRegNum {
      i.Event().OnData(s, t, r.reg[slot.port])
    } else {
      d := bus.IntData{D: rand.Int() % 999}
      i.Event().OnData(s, t, &d)
    }
  }
  return nil
}


func (r *random_bus) SendCtrl(s bus.Slot, d bus.DataWrap, t *time.Time) error {
  slot := s.(*bus_random_sl)
  if slot.port < RealRegNum {
    r.reg[slot.port] = d
  }
  return nil
}


func (r *random_bus) Stop(i bus.BusReal) {
  i.Log("总线停止")
}
