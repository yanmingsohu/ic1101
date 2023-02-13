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
	"fmt"
	"ic1101/src/bus"
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


const _modbus_format = "N%xS%xR%xT%xL%x"
//
// slot 格式: N{nn}S{cc}R{aaaa}T{t}L{l}
//  N = 16 进制, 控制码
//  C = 16 进制, 从机地址
//  A = 16 进制, 数据地址
//  T = 16 进制, 数据类型
//  L = 字节序
//
func _parse_modbus_slot(s string) (bus.Slot, error) {
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

