package bus

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/goburrow/modbus"
)

const (
  MB_r_coils = 0x01
  MB_r_discreta_inputs = 0x02
  MB_r_holding_registers = 0x03
  MB_r_input_registers = 0x04
  MB_w_coil = 0x05
  MB_w_register = 0x06
  MB_wm_coils = 0x0F
  MB_wm_registers = 0x10
  MB_r_file_record = 0x14
  MB_w_file_record = 0x15
  MB_rwm_registers = 0x17
  MB_r_fifo = 0x18
)

const (
  _ = iota
  mb_dt_uint8  byte = iota
  mb_dt_int8
  mb_dt_uint16
  mb_dt_int16
  mb_dt_uint32
  mb_dt_int32
  mb_dt_float32
  mb_dt_bool
)


func init() {
  InstallBus("modbus", &bus_modbus_ct{})
}


type bus_modbus_ct struct {
}


func (*bus_modbus_ct) Name() string {
  return "MODBUS 总线"
}


func (*bus_modbus_ct) Create(i *BusInfo) (Bus, error) {
  return &modebus_impl{}, nil
}


// 接受任何字符串作为 slot
func (*bus_modbus_ct) ParseSlot(s string) (Slot, error) {
  return _parse_modbus_slot(s)
}


func (*bus_modbus_ct) SlotDesc(s string) (string, error) {
  slot, err := _parse_modbus_slot(s)
  if err != nil {
    return "", err
  }
  return slot.Desc(), nil
}


//
// modbus uri 格式:
//   tcp://[host][:port]
//   rtu://[/path]
//   asc://[/path]
//   dtu://[host][:port][/dtu-type]
//
func (*bus_modbus_ct) ParseURI(uri string) (*url.URL, error) {
  u, err := url.Parse(uri)
  if err != nil {
    return nil, err
  }
  switch u.Scheme {
  case "tcp":
  case "rtu":
  case "asc":
  case "dtu":
  default:
    return nil, errors.New("scheme 只支持 tcp://, rtu://, asc://, dtu:// 模式");
  }
  return u, nil
}


type modebus_impl struct {
  h   modbus.ClientHandler
  c   modbus.Client
  sid *byte
}


func (r *modebus_impl) start(i *BusInfo) error {
  switch i.uri.Scheme {
  case "tcp":
    h := modbus.NewTCPClientHandler(i.uri.Host)
    r.sid = &h.SlaveId
    r.h = h
  case "rtu":
    h := modbus.NewRTUClientHandler(i.uri.Path)
    r.sid = &h.SlaveId
    r.h = h
  case "asc":
    h := modbus.NewASCIIClientHandler(i.uri.Path)
    r.sid = &h.SlaveId
    r.h = h
  case "dtu":

  default:
    return errors.New("无效的scheme")
  }
  r.c = modbus.NewClient(r.h)
  i.Log("总线启动")
  return nil
}


func (r *modebus_impl) sync_data(i *BusInfo, t *time.Time) error {
  for _, s := range i.datas {
    ms := s.(*modbus_slot)
    *r.sid = ms.c
    var b []byte
    var err error

    switch ms.n {
    case MB_r_coils:
      b, err = r.c.ReadCoils(ms.a, 1)

    case MB_r_discreta_inputs:
      b, err = r.c.ReadDiscreteInputs(ms.a, 1)

    case MB_r_holding_registers:
      b, err = r.c.ReadHoldingRegisters(ms.a, 1)

    case MB_r_input_registers:
      b, err = r.c.ReadInputRegisters(ms.a, 1)

    // case MB_r_fifo:
    //   b, err = r.c.ReadFIFOQueue(ms.a)
    default:
      return errors.New("无效的操作码")
    }

    if err != nil {
      return err
    }
    d := IntData{int(b[0])}
    i.event.OnData(s, t, &d)
  }
  return nil
}


func (r *modebus_impl) send_ctrl(s Slot, d DataWrap, t *time.Time) error {
  ms := s.(*modbus_slot)
  *r.sid = ms.c

  switch ms.n {
  case MB_w_coil:
    r.c.WriteSingleCoil(ms.a, uint16(d.Int()))

  case MB_w_register:
    r.c.WriteSingleRegister(ms.a, uint16(d.Int()))
    
  default:
    return errors.New("不支持的操作")
  }
  return nil
}


func (r *modebus_impl) stop(i *BusInfo) {
  switch r.h.(type) {
  case io.Closer:
    r.h.(io.Closer).Close()
  }
  i.Log("总线停止")
}


//
// slot 格式: !NN@CC$AAAA&T
//  N = 16 进制, 控制码
//  C = 16 进制, 从机地址
//  A = 16 进制, 数据地址
//  T = 16 进制, 数据类型
//
func _parse_modbus_slot(s string) (Slot, error) {
  m := modbus_slot{}
  n, err := fmt.Sscanf(s, "!%x@%x$%x", &m.n, &m.c, &m.a)
  if err != nil {
    return nil, err
  }
  if n != 3 {
    return nil, errors.New("无效格式")
  }
  return &m, nil
}


type modbus_slot struct {
  // 控制码
  n  byte
  // 从机地址
  c  byte
  // 数据地址
  a  uint16  
  // 数据类型
  t  byte
}


func (m *modbus_slot) String() string {
  return fmt.Sprintf("!%x@%x$%x&%x", m.n, m.c, m.a, m.t)
}


func (m *modbus_slot) Desc() string {
  var name string
  switch m.n {
  case MB_r_coils:
    name = "读线圈"
  case MB_r_discreta_inputs:
    name = "读离散"
  case MB_r_holding_registers:
    name = "读保持寄存器"
  case MB_r_input_registers:
    name = "读输入寄存器"
  case MB_w_coil:
    name = "写线圈"
  case MB_w_register:
    name = "写寄存器"
  case MB_r_fifo:
    name = "读队列"
  default:
    return "不支持的操作"
  }
  return fmt.Sprintf("%s 从机:%d 地址:%d", name, m.c, m.a)
}


func (m *modbus_slot) Type() SlotType {
  switch m.n {
  case MB_r_coils:
    return SlotData
  case MB_r_discreta_inputs:
    return SlotData
  case MB_r_holding_registers:
    return SlotData
  case MB_r_input_registers:
    return SlotData
  case MB_w_coil:
    return SlotCtrl
  case MB_w_register:
    return SlotCtrl
  case MB_r_fifo:
    return SlotData
  }
  return SlotInvaild
}