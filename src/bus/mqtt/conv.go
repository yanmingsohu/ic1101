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
package bus_mqtt

import (
	"encoding/binary"
	"errors"
	"ic1101/src/bus"
	"math"
)

type mq_conv_dw func([]byte) bus.DataWrap


func get_conv_mq2dw(t byte) (mq_conv_dw, error) {
  switch t {
  case st_uint8:
    return md_st_uint8, nil

  case st_int8:
    return md_st_int8, nil

  case st_uint16:
    return md_st_uint16, nil

  case st_int16:
    return md_st_int16, nil

  case st_uint32:
    return md_st_uint32, nil

  case st_int32:
    return md_st_int32, nil

  case st_uint64:
    return md_st_uint64, nil

  case st_int64:
    return md_st_int64, nil

  case st_float32:
    return md_st_float32, nil

  case st_float64:
    return md_st_float64, nil

  case st_string:
    return md_st_string, nil
  }
  return nil, errors.New("无效数据类型")
}


func conv_dw2mq(t byte, d bus.DataWrap) []byte {
  switch t {
  case st_uint8: fallthrough
  case st_int8:
    return []byte{uint8(d.Int())}

  case st_uint16: fallthrough
  case st_int16:
    b := make([]byte, 2)
    binary.BigEndian.PutUint16(b, uint16(d.Int()))
    return b

  case st_uint32: fallthrough
  case st_int32:
    b := make([]byte, 4)
    binary.BigEndian.PutUint32(b, uint32(d.Int()))
    return b

  case st_uint64: fallthrough
  case st_int64:
    b := make([]byte, 8)
    binary.BigEndian.PutUint64(b, uint64(d.Int64()))
    return b

  case st_float32:
    bits := math.Float32bits(d.Float())
    b := make([]byte, 4)
    binary.BigEndian.PutUint32(b, bits)
    return b

  case st_float64:
    bits := math.Float64bits(d.Float64())
    b := make([]byte, 4)
    binary.BigEndian.PutUint64(b, bits)
    return b

  case st_string:
    return []byte(d.String())
  }
  return []byte{0}
}


func md_st_uint8(b []byte) bus.DataWrap {
  r := bus.UInt64Data{D:0}
  if len(b) >= 1 {
    r.D = uint64(b[0])
  }
  return &r
}


func md_st_int8(b []byte) bus.DataWrap {
  r := bus.Int64Data{D:0}
  if len(b) >= 1 {
    r.D = int64(int8(b[0]))
  }
  return &r
}


func md_st_uint16(b []byte) bus.DataWrap {
  r := bus.UInt64Data{D:0}
  if len(b) >= 2 {
    r.D = uint64(binary.BigEndian.Uint16(b))
  }
  return &r
}


func md_st_int16(b []byte) bus.DataWrap {
  r := bus.Int64Data{D:0}
  if len(b) >= 2 {
    r.D = int64(int16(binary.BigEndian.Uint16(b)))
  }
  return &r
}


func md_st_uint32(b []byte) bus.DataWrap {
  r := bus.UInt64Data{D:0}
  if len(b) >= 4 {
    r.D = uint64(binary.BigEndian.Uint32(b))
  }
  return &r
}


func md_st_int32(b []byte) bus.DataWrap {
  r := bus.Int64Data{D:0}
  if len(b) >= 4 {
    r.D = int64(int32(binary.BigEndian.Uint32(b)))
  }
  return &r
}


func md_st_uint64(b []byte) bus.DataWrap {
  r := bus.UInt64Data{D:0}
  if len(b) >= 8 {
    r.D = binary.BigEndian.Uint64(b)
  }
  return &r
}


func md_st_int64(b []byte) bus.DataWrap {
  r := bus.Int64Data{D:0}
  if len(b) >= 8 {
    r.D = int64(binary.BigEndian.Uint64(b))
  }
  return &r
}


func md_st_float32(b []byte) bus.DataWrap {
  r := bus.FloatData{D:0}
  if len(b) >= 4 {
    i := binary.BigEndian.Uint32(b)
    r.D = math.Float32frombits(i)
  }
  return &r
}


func md_st_float64(b []byte) bus.DataWrap {
  r := bus.Float64Data{D:0}
  if len(b) >= 8 {
    i := binary.BigEndian.Uint64(b)
    r.D = math.Float64frombits(i)
  }
  return &r
}


func md_st_string(b []byte) bus.DataWrap {
  return &bus.StringData{D: string(b)}
}