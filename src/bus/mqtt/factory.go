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
	"errors"
	"ic1101/src/bus"
	"net/url"
)


func init() {
  bus.InstallBus("mqtt", &mqtt_ct{})
}


type mqtt_ct struct {
}


func (*mqtt_ct) Name() string {
  return "MQTT 客户端"
}


func (*mqtt_ct) Create(i bus.BusReal) (bus.Bus, error) {
  return &bus_impl{}, nil
}


func (*mqtt_ct) ParseSlot(s string) (bus.Slot, error) {
  return parse_slot(s)
}


func (*mqtt_ct) SlotDesc(s string) (string, error) {
  slot, err := parse_slot(s)
  if err != nil {
    return "", err
  }
  return slot.Desc(), nil
}


func (*mqtt_ct) ParseURI(uri string) (*url.URL, error) {
  u, err := url.Parse(uri)
  if err != nil {
    return nil, err
  }
  switch u.Scheme {
  case "tcp":
  case "ws":
  case "wss":
  case "ssl":
  default:
    return nil, errors.New("无效的 Scheme [tcp/ssl/ws/wss]")
  }
  return u, nil
}