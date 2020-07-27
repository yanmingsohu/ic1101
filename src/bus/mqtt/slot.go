package bus_mqtt


import (
	"errors"
	"fmt"
	"ic1101/src/bus"
)


//
// slot 格式: [D|C]#{主题} {数据类型}/{偏移}/{QOS}
// * scanf 必须用空格分隔字符串
//
const slot_fmt = "%c#%s %x/%x/%d"

const (
  _ = iota
  st_uint8 byte = iota
  st_int8 
  st_uint16 
  st_int16
  st_uint32
  st_int32
  st_uint64
  st_int64
  st_float32
  st_float64
  st_string

  st_invaild //保持最后
)


func parse_slot(str string) (bus.Slot, error) {
  s := slot_impl{}
  var c uint8
  n, err := fmt.Sscanf(str, slot_fmt, &c, &s.topic, &s.data_type, &s.offset, &s.q)
  if err != nil {
    return nil, err
  }

  if n != 5 {
    return nil, errors.New("无效格式")
  }

  if s.data_type <= 0 || s.data_type >= st_invaild {
    return nil, errors.New("无效数据类型")
  }

  if c == 'D' {
    s.st = bus.SlotData
  } else if c == 'C' {
    s.st = bus.SlotCtrl
  } else {
    return nil, errors.New("无效槽类型")
  }
  
  if s.q < 0 || s.q > 2 {
    return nil, errors.New("无效QOS")
  }
  return &s, nil
}


type slot_impl struct {
  // topic name
  topic string
  // 数据类型
  data_type byte
  // 偏移
  offset uint16
  // 槽类型
  st bus.SlotType
  // QOS
  q byte
}


func (s *slot_impl) String() string {
  var c uint8
  if s.st == bus.SlotData {
    c = 'D'
  } else {
    c = 'C'
  }
  return fmt.Sprintf(slot_fmt, c, s.topic, s.data_type, s.offset, s.q)
}


func (s *slot_impl) Desc() string {
  return fmt.Sprintf("主题 %s 类型 %d 偏移 %d", s.topic, s.data_type, s.offset)
}


func (s *slot_impl) Type() bus.SlotType {
  return s.st
}