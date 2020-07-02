package bus

import "errors"

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