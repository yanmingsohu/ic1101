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
