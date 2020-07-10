package dtu

import (
	"errors"
	"log"
	"net"
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
  if url.Scheme != "dtu" {
    return nil, errors.New("不是 dtu URL")
  }
  
  factName := url.EscapedPath()
  if factName[0] == '/' {
    factName = factName[1:]
  }

  fact, has := _dtu_factorys[factName]
  if has {
    return fact.New(url, handle)
  }
  return nil, errors.New("DTU 不存在 "+ factName)
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


//
// 创建 DTU 实例的工厂
//
type Fact interface {
  //
  // 返回 DTU 实例, 或创建失败的原因
  //
  New(url *url.URL, handle Event) (Impl, error)
  //
  // 该 dtu 的类型名, 使用该名称注册到 dtu 注册表
  //
  Name() string
  //
  // 对 dtu 的描述 如: 型号, 版本, 厂家
  //
  Desc() string
}


//
// 与 DTU 之间交换数据, 一个 DTU 的实现可以与一组物理 DTU 通信,
// 此时 DTU 处于 tcp/udp 服务器模式, 使用 Context 来确定对端.
//
type Impl interface {
  //
  // 返回上下文, 如果上下文关闭返回错误
  //
  GetContext(id int) (Context, error)
}


//
// 上下文用于确定, 一个对端(客户端)实例的连接, 该对象可以安全的保存或重用,
// 该对象生命期总是短于 Impl.
//
type Context interface {
  //
  // 返回该上下文的唯一标识
  //
  Id() (int)
  //
  // 返回上下文参数, 不同的 DTU 实现会有不同的参数定义.
  //
  Get(name string) (interface{}, error)
  //
  // 允许在上下文绑定变量, 之后用 Get() 可以取得该变量.
  //
  Bind(name string, data interface{})
  //
  // 关闭这个上下文, 实现通常会关闭底层网络连接, 重复调用该方法是安全的.
  //
  Close()
  //
  // 对于该函数的使用方, DTU 中的数据转换是透明的, 
  // 用户只关心与设备之间的原始数据格式.
  // 如果 dtu 不支持该方法则总是返回错误,
  // Conn.Close() 方法与 Context.Close() 有相同的效果
  //
  GetConn() (net.Conn, error)
}


//
// 接受 dtu 消息
//
type Event interface {
  //
  // 创建新的上下文, 通常是因为一个新的客户端发起了请求.
  //
  NewContext(ctx Context)
}