package bus_random

import (
	"ic1101/src/bus"
	"net/url"
)


func init() {
  bus.InstallBus("random", &bus_random_ct{})
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