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
package bus_random

import (
	"errors"
	"fmt"
	"ic1101/src/bus"
	"strconv"
)


type bus_random_sl struct {
  port int
  tp bus.SlotType
}


func (s *bus_random_sl) String() string {
  if s.tp == bus.SlotData {
    return "D#"+ strconv.Itoa(s.port)
  } else {
    return "C#"+ strconv.Itoa(s.port)
  }
}


func (s *bus_random_sl) Desc() string {
  if s.port < RealRegNum {
    if s.tp == bus.SlotData {
      return "虚拟寄存器 "+ strconv.Itoa(s.port)
    } else {
      return "虚拟控制 "+ strconv.Itoa(s.port)
    }
  } else {
    if s.tp == bus.SlotData {
      return "随机数据 "+ strconv.Itoa(s.port)
    } else {
      return "空控制 "+ strconv.Itoa(s.port)
    }
  }
}


func (s *bus_random_sl) Type() bus.SlotType {
  return s.tp
}


func _parse_random_slot(s string) (*bus_random_sl, error) {
  var t int
  var port int
  var tp bus.SlotType

  n, err := fmt.Sscanf(s, "%c#%d", &t, &port)
  if err != nil {
    return nil, err
  }
  if n != 2 {
    return nil, errors.New("无效的slot格式")
  }
  switch (t) {
  case 'D':
    tp = bus.SlotData
  case 'C':
    tp = bus.SlotCtrl
  default:
    return nil, errors.New("无效的类型字符")
  }
  return &bus_random_sl{port, tp}, nil
}