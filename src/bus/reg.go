package bus

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