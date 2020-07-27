package bus_random

import (
	"ic1101/src/bus"
	"math/rand"
	"time"
)

const RealRegNum = 0x0F;


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
