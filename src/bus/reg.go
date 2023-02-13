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
	"log"
)

//
// 总线类型注册表, 所有可用的总线注册到这里
//
var bus_type_register = map[string]BusCreator{
}


func InstallBus(id string, ct BusCreator) {
  if _, has := bus_type_register[id]; has {
    panic("总线已经被注册 "+ id)
  }
  bus_type_register[id] = ct
  log.Println("BUS reg:", id)
}


func GetTypes() map[string]string {
  ret := map[string]string{}
  for id, ct := range bus_type_register {
    ret[id] = ct.Name()
  }
  return ret
}


func HasTypeName(name string) bool {
  _, has := bus_type_register[name]
  return has
}


//
// 返回对应总线类型的数据槽解析器
//
func GetSlotParser(typeName string) (SlotParser, error) {
  ct, has := bus_type_register[typeName]
  if !has {
    return nil, errors.New("不存在的总线类型 "+ typeName)
  }
  return ct, nil
}