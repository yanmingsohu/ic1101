package bus_random

import (
	"errors"
	"fmt"
	"ic1101/src/bus"
	"strconv"
)


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