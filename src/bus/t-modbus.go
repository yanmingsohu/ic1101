package bus

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/simonvetter/modbus"
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
  mb_dt_uint16 byte = iota
  mb_dt_int16
  mb_dt_uint32
  mb_dt_int32
  mb_dt_uint64
  mb_dt_int64
  mb_dt_float32
  mb_dt_float64
)

const (
  _ = iota
  // BIG_ENDIAN HIGH_WORD_FIRST
  mb_bit_big_endian byte = iota
  // BIG_ENDIAN LOW_WORD_FIRST
  mb_bit_big_endian_swap
  // LITTLE_ENDIAN LOW_WORD_FIRST
  mb_bit_little_endian
  // LITTLE_ENDIAN HIGH_WORD_FIRST
  mb_bit_little_endian_swap
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
  if i.uri.Scheme == "dtu" {
    panic("未实现")
  }
  return &modbus_s_impl{}, nil
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
  case "rtuovertcp":
  case "dtu":
  default:
    return nil, errors.New("scheme 必须是: tcp://, rtu://, rtuovertcp://, dtu://");
  }
  return u, nil
}


type modbus_s_impl struct {
  c   MC
  sid *byte
}


func (r *modbus_s_impl) start(i *BusInfo) (err error) {
  var c *modbus.ModbusClient

  switch i.uri.Scheme {
  case "tcp":
    c, err = modbus.NewClient(&modbus.ClientConfiguration{
      URL:      i.uri.String(),
      Timeout:  1 * time.Second,
    })

  case "rtu":
    c, err = modbus.NewClient(&modbus.ClientConfiguration{
      URL:      i.uri.String(),
      Speed:    9600,
      DataBits: 8,
      Parity:   modbus.PARITY_NONE,
      StopBits: 2,
      Timeout:  300 * time.Millisecond,
    })

  case "rtuovertcp":
    c, err = modbus.NewClient(&modbus.ClientConfiguration{
      URL:      i.uri.String(),
      Timeout:  1 * time.Second,
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
  i.Log("总线启动, " + i.uri.String())
  return
}


func (r *modbus_s_impl) sync_data(i *BusInfo, t *time.Time) (err error) {
  for _, s := range i.datas {
    ms := s.(*modbus_slot)

    r.c.SetUnitId(ms.c)
    r.c.setMode(ms.l)
    d, err := r.c.read(ms)
    if err != nil {
      return err
    }

    i.event.OnData(s, t, d)
  }
  return nil
}


func (r *modbus_s_impl) send_ctrl(_s Slot, d DataWrap, t *time.Time) error {
  s := _s.(*modbus_slot)
  r.c.SetUnitId(s.c)
  r.c.setMode(s.l)
  return r.c.send(s, d)
}


func (r *modbus_s_impl) stop(i *BusInfo) {
  r.c.Close()
  i.Log("总线停止")
}


type MC struct {
  *modbus.ModbusClient
}


func (c *MC) setMode(b byte) {
  switch b {
  case mb_bit_big_endian:
    c.SetEncoding(modbus.BIG_ENDIAN, modbus.HIGH_WORD_FIRST)

  case mb_bit_big_endian_swap:
    c.SetEncoding(modbus.BIG_ENDIAN, modbus.LOW_WORD_FIRST)

  case mb_bit_little_endian:
    c.SetEncoding(modbus.LITTLE_ENDIAN, modbus.LOW_WORD_FIRST)

  case mb_bit_little_endian_swap:
    c.SetEncoding(modbus.LITTLE_ENDIAN, modbus.HIGH_WORD_FIRST)
  }
}


func (c *MC) send(s *modbus_slot, d DataWrap) error {
  switch s.n {
  case MB_w_coil:
    return c.WriteCoil(s.a, d.Bool())

  case MB_w_register:

    switch s.l {
    case mb_dt_uint16:
      return c.WriteRegister(s.a, uint16(d.Int()))

    case mb_dt_int16:
      return c.WriteRegister(s.a, uint16(d.Int()))

    case mb_dt_uint32:
      return c.WriteUint32(s.a, uint32(d.Int()))

    case mb_dt_int32:
      return c.WriteUint32(s.a, uint32(d.Int()))

    case mb_dt_uint64:
      return c.WriteUint64(s.a, uint64(d.Int64()))

    case mb_dt_int64:
      return c.WriteUint64(s.a, uint64(d.Int64()))

    case mb_dt_float32:
      return c.WriteFloat32(s.a, d.Float())

    case mb_dt_float64:
      return c.WriteFloat64(s.a, d.Float64())

    default:
      return errors.New("无效的数据长度")
    }
    
  default:
    return errors.New("不支持的操作")
  }
}


func (c *MC) read(s *modbus_slot) (DataWrap, error) {
  switch s.n {
  case MB_r_coils:
    v, err := c.ReadCoil(s.a)
    if err != nil {
      return nil, err
    }
    return &BoolData{v}, nil

  case MB_r_discreta_inputs:
    v, err := c.ReadDiscreteInput(s.a)
    if err != nil {
      return nil, err
    }
    return &BoolData{v}, nil

  case MB_r_holding_registers:
    v, err := c.read_reg(s, modbus.HOLDING_REGISTER)
    if err != nil {
      return nil, err
    }
    return v, nil

  case MB_r_input_registers:
    v, err := c.read_reg(s, modbus.INPUT_REGISTER)
    if err != nil {
      return nil, err
    }
    return v, nil

  default:
    return nil, errors.New("无效的操作码")
  }
}


func (c *MC) read_reg(s *modbus_slot, t modbus.RegType) (DataWrap, error) {
  switch s.t {
  case mb_dt_uint16:
    v, err := c.ReadRegister(s.a, t)
    if err != nil {
      return nil, err
    }
    return &UInt64Data{uint64(v)}, nil

  case mb_dt_int16:
    v, err := c.ReadRegister(s.a, t)
    if err != nil {
      return nil, err
    }
    return &Int64Data{int64(int16(v))}, nil

  case mb_dt_uint32:
    v, err := c.ReadUint32(s.a, t)
    if err != nil {
      return nil, err
    }
    return &UInt64Data{uint64(v)}, nil

  case mb_dt_int32:
    v, err := c.ReadUint32(s.a, t)
    if err != nil {
      return nil, err
    }
    return &Int64Data{int64(int32(v))}, nil

  case mb_dt_uint64:
    v, err := c.ReadUint64(s.a, t)
    if err != nil {
      return nil, err
    }
    return &UInt64Data{v}, nil
    
  case mb_dt_int64:
    v, err := c.ReadUint64(s.a, t)
    if err != nil {
      return nil, err
    }
    return &Int64Data{int64(v)}, nil

  case mb_dt_float32:
    v, err := c.ReadFloat32(s.a, t)
    if err != nil {
      return nil, err
    }
    return &FloatData{v}, nil

  case mb_dt_float64:
    v, err := c.ReadFloat64(s.a, t)
    if err != nil {
      return nil, err
    }
    return &Float64Data{v}, nil
  }
  return nil, errors.New("无效的数据长度")
}


const _modbus_format = "N%xS%xR%xT%xL%x"
//
// slot 格式: N{nn}S{cc}R{aaaa}T{t}L{l}
//  N = 16 进制, 控制码
//  C = 16 进制, 从机地址
//  A = 16 进制, 数据地址
//  T = 16 进制, 数据类型
//  L = 字节序
//
func _parse_modbus_slot(s string) (Slot, error) {
  m := modbus_slot{}
  n, err := fmt.Sscanf(s, _modbus_format, &m.n, &m.c, &m.a, &m.t, &m.l)
  if err != nil {
    return nil, err
  }
  if n != 5 {
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
  // 字节序
  l  byte
}


func (m *modbus_slot) String() string {
  return fmt.Sprintf(_modbus_format, m.n, m.c, m.a, m.t, m.l)
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