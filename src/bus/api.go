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
package bus

import (
	"errors"
	"net/url"
	"sync"
)

//
// 运行中的总线实例
//
var _bus_mutex = map[string]*BusInfo{}
var _api_mutex = new(sync.RWMutex)
const MaxLogCount = 20


//
// 对总线使用者隐藏
// 实现 BusReal
//
type bus_real_impl struct {
  *BusInfo
}


func (b *bus_real_impl) URL() *url.URL {
  return b.uri
}


func (b *bus_real_impl) Event() BusEvent {
  return b.event
}


func (b *bus_real_impl) Datas() []Slot {
  return b.datas[:]
}


//
// 返回总线状态
//
func GetBusState(id string) BusState {
  _api_mutex.RLock()
  defer _api_mutex.RUnlock()

  if info, has := _bus_mutex[id]; has {
    return info.State()
  }
  return BusStateStop
}


//
// 返回运行中的总线
//
func GetBus(id string) (*BusInfo, error) {
  _api_mutex.RLock()
  defer _api_mutex.RUnlock()

  if info, has := _bus_mutex[id]; has {
    return info, nil
  }
  return nil, errors.New(id +" 引用的总线没有运行")
}


//
// 启动总线, 成功返回 nil, 否则返回错误
//
func StartBus(info *BusInfo) (error) {
  _api_mutex.Lock()
  defer _api_mutex.Unlock()
  
  if _, has := _bus_mutex[info.id]; has {
    return errors.New("总线已经启动 "+ info.id)
  }

  ct, has := bus_type_register[info.typeName]
  if !has {
    return errors.New("无效的总线类型 "+ info.typeName)
  }

  real := &bus_real_impl{info}
  bus, err := ct.Create(real)
  if err != nil {
    return err
  }
  if err := info.start(bus); err != nil {
    return err
  }
  _bus_mutex[info.id] = info
  return nil
}


//
// 停止总线, 成功返回 nil, 否则返回 error
//
func StopBus(id string) error {
  _api_mutex.Lock()
  defer _api_mutex.Unlock()

  info, has := _bus_mutex[id]
  if !has {
    return errors.New("总线没有启动 "+ id)
  }
  info.stop()
  delete(_bus_mutex, id)
  return nil
}