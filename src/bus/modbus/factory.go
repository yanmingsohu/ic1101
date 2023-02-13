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
	"ic1101/src/dtu"
	"net/url"
)


func init() {
  bus.InstallBus("modbus", &bus_modbus_ct{})
}


type bus_modbus_ct struct {
}


func (*bus_modbus_ct) Name() string {
  return "MODBUS 总线"
}


func (*bus_modbus_ct) Create(i bus.BusReal) (bus.Bus, error) {
  if i.URL().Scheme == "dtu" {
    return &tcp_server{}, nil
  }
  return &modbus_s_impl{}, nil
}


// 接受任何字符串作为 slot
func (*bus_modbus_ct) ParseSlot(s string) (bus.Slot, error) {
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
//   rtuovertcp://[/path]
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
    if _, err = dtu.CheckUrl(u); err != nil {
      return nil, err
    }
  default:
    return nil, errors.New("scheme 必须是: tcp://, rtu://, rtuovertcp://, dtu://");
  }
  return u, nil
}