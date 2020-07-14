package bus_modbus

import (
	"errors"
	"ic1101/src/bus"
	"time"

	"github.com/simonvetter/modbus"
)



type modbus_s_impl struct {
  c   MC
  sid *byte
}


func (r *modbus_s_impl) Start(i bus.BusReal) (err error) {
  var c *modbus.ModbusClient

  switch i.URL().Scheme {
  case "tcp":
    c, err = modbus.NewClient(&modbus.ClientConfiguration{
      URL:      i.URL().String(),
      Timeout:  10 * time.Second,
    })

  case "rtu":
    c, err = modbus.NewClient(&modbus.ClientConfiguration{
      URL:      i.URL().String(),
      Speed:    9600,
      DataBits: 8,
      Parity:   modbus.PARITY_NONE,
      StopBits: 2,
      Timeout:  300 * time.Millisecond,
    })

  case "rtuovertcp":
    c, err = modbus.NewClient(&modbus.ClientConfiguration{
      URL:      i.URL().String(),
      Timeout:  10 * time.Second,
    })

  default:
    return errors.New("无效的scheme")
  }

  if err != nil {
    return
  }
  
  r.c = MC{c}
  if err := r.c.Open(); err != nil {
    return err
  }
  i.Log("总线启动, " + i.URL().String())
  return
}


func (r *modbus_s_impl) SyncData(i bus.BusReal, t *time.Time) (err error) {
  for _, s := range i.Datas() {
    ms := s.(*modbus_slot)

    r.c.SetUnitId(ms.c)
    r.c.setMode(ms.l)
    d, err := r.c.read(ms)
    if err != nil {
      return err
    }

    i.Event().OnData(s, t, d)
  }
  return nil
}


func (r *modbus_s_impl) SendCtrl(_s bus.Slot, d bus.DataWrap, t *time.Time) error {
  s := _s.(*modbus_slot)
  r.c.SetUnitId(s.c)
  r.c.setMode(s.l)
  return r.c.send(s, d)
}


func (r *modbus_s_impl) Stop(i bus.BusReal) {
  r.c.Close()
  i.Log("总线停止")
}

