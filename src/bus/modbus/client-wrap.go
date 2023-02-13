/**
 *  Copyright 2023 Jing Yanming
 * 
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */
package bus_modbus

import (
	"errors"
	"ic1101/src/bus"

	"github.com/simonvetter/modbus"
)


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


func (c *MC) send(s *modbus_slot, d bus.DataWrap) error {
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


func (c *MC) read(s *modbus_slot) (bus.DataWrap, error) {
  switch s.n {
  case MB_r_coils:
    v, err := c.ReadCoil(s.a)
    if err != nil {
      return nil, err
    }
    return &bus.BoolData{D: v}, nil

  case MB_r_discreta_inputs:
    v, err := c.ReadDiscreteInput(s.a)
    if err != nil {
      return nil, err
    }
    return &bus.BoolData{D: v}, nil

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


func (c *MC) read_reg(s *modbus_slot, t modbus.RegType) (bus.DataWrap, error) {
  switch s.t {
  case mb_dt_uint16:
    v, err := c.ReadRegister(s.a, t)
    if err != nil {
      return nil, err
    }
    return &bus.UInt64Data{D: uint64(v)}, nil

  case mb_dt_int16:
    v, err := c.ReadRegister(s.a, t)
    if err != nil {
      return nil, err
    }
    return &bus.Int64Data{D: int64(int16(v))}, nil

  case mb_dt_uint32:
    v, err := c.ReadUint32(s.a, t)
    if err != nil {
      return nil, err
    }
    return &bus.UInt64Data{D: uint64(v)}, nil

  case mb_dt_int32:
    v, err := c.ReadUint32(s.a, t)
    if err != nil {
      return nil, err
    }
    return &bus.Int64Data{D: int64(int32(v))}, nil

  case mb_dt_uint64:
    v, err := c.ReadUint64(s.a, t)
    if err != nil {
      return nil, err
    }
    return &bus.UInt64Data{D: v}, nil
    
  case mb_dt_int64:
    v, err := c.ReadUint64(s.a, t)
    if err != nil {
      return nil, err
    }
    return &bus.Int64Data{D: int64(v)}, nil

  case mb_dt_float32:
    v, err := c.ReadFloat32(s.a, t)
    if err != nil {
      return nil, err
    }
    return &bus.FloatData{D: v}, nil

  case mb_dt_float64:
    v, err := c.ReadFloat64(s.a, t)
    if err != nil {
      return nil, err
    }
    return &bus.Float64Data{D: v}, nil
  }
  return nil, errors.New("无效的数据长度")
}