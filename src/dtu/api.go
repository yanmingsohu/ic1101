package dtu

import (
	"errors"
	"log"
	"net/url"
)


// 不要直接引用
var _dtu_factorys = make(map[string]Fact)


//
// 解析 url 并返回 dtu 实现对象
// dtu url 格式:
//   dtu://[host][:port][/dtu-type][?parameaters]
//
func GetDtu(url *url.URL, handle Event) (Impl, error) {
  if handle == nil {
    return nil, errors.New("事件处理不能为空")
  }
  fact, err := CheckUrl(url)
  if err != nil {
    return nil, err
  }
  return fact.New(url, handle)
}


//
// 检查 url 的基础参数, 只有在绝对可以生成有效 DTU 的时候才返回 DTU 工厂
//
func CheckUrl(url *url.URL) (Fact, error) {
  if url.Scheme != "dtu" {
    return nil, errors.New("不是 dtu URL")
  }
  
  factName := url.EscapedPath()
  if factName == "" {
    return nil, errors.New("必须选择 DTU 类型")
  }
  if factName[0] == '/' {
    factName = factName[1:]
  } else {
    return nil, errors.New("必须选择 DTU 类型")
  }
  f, has := _dtu_factorys[factName]
  if !has {
    return nil, errors.New("DTU 不存在 "+ factName)
  }
  return f, nil
}


//
// 注册一个 dtu 工厂, 任何错误都会导致 panic, 如命名冲突
//
func RegFact(f Fact) {
  if _, has := _dtu_factorys[f.Name()]; has {
    panic(errors.New("工厂名称冲突" + f.Name()))
  }
  _dtu_factorys[f.Name()] = f
  log.Println("DTU reg:", f.Name(), f.Desc())
}