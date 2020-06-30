package bus

import (
	"math/rand"
	"time"
)


func init() {
  InstallBus("random", &bus_random_ct{})
}


type bus_random_ct struct {
}


func (*bus_random_ct) Name() string {
  return "随机数(测试)"
}


func (*bus_random_ct) Create(i *BusInfo) (Bus, error) {
  return &random_bus{i, BusStateStartup}, nil
}


// 接受任何字符串作为 slot
func (*bus_random_ct) ParseSlot(s string) (Slot, error) {
  return &bus_random_sl{s, SlotData}, nil
}


func (*bus_random_ct) SlotDesc(s string) (string, error) {
  return "虚拟端口 "+ s, nil
}


type bus_random_sl struct {
  id string
  tp SlotType
}


func (s *bus_random_sl) String() string {
  return s.id
}


func (s *bus_random_sl) Type() SlotType {
  return s.tp
}


type random_bus struct {
  info    *BusInfo
  state   BusState
}


func (r *random_bus) start() error {
  r.state = BusStateSleep
  r.info.Tk.Start(func() {
    r.state = BusStateTask
    t := time.Now()
    
    for _, s := range r.info.SlotConf {
      d := IntData{rand.Int()}
      r.info.Recv.OnData(s, &t, &d)
    }

    r.state = BusStateSleep
  }, func() {
    r.stop()
  })
  return nil
}


func (r *random_bus) stop() {
  r.state = BusStateStop
}


func (*random_bus) SendCtrl(s Slot, d DataWrap) error {
  return nil
}


func (r *random_bus) State() BusState {
  return r.state
}
