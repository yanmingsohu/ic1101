package bus_modbus

import (
	"fmt"
	"ic1101/src/bus"
)


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


func (m *modbus_slot) Type() bus.SlotType {
  switch m.n {
  case MB_r_coils:
    return bus.SlotData
  case MB_r_discreta_inputs:
    return bus.SlotData
  case MB_r_holding_registers:
    return bus.SlotData
  case MB_r_input_registers:
    return bus.SlotData
  case MB_w_coil:
    return bus.SlotCtrl
  case MB_w_register:
    return bus.SlotCtrl
  case MB_r_fifo:
    return bus.SlotData
  }
  return bus.SlotInvaild
}


//
// 为错误提供更详尽的信息
//
func (m *modbus_slot) ErrInfo(err interface{}) string {
  return fmt.Sprintf("%s, 从机 %d, 地址 %d", err, m.c, m.a)
}


//
// 逻辑地址是 从机 + 寄存器地址
//
func (m *modbus_slot) LogicAddr() uint64 {
  return (uint64(m.c) << 16) + uint64(m.a)
}