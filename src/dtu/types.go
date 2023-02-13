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
package dtu

import (
	"net"
	"net/url"
)


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
// 该对象是线程安全的
//
type Impl interface {
  //
  // 返回上下文, 如果上下文关闭返回错误
  //
  GetContext(id int) (Context, error)
  //
  // 停止 DTU 所有任务
  //
  Stop()
  //
  // 关闭上下文, 并释放内存
  //
  CloseContext(id int) error
}


//
// 上下文用于确定, 一个对端(客户端)实例的连接, 该对象可以安全的保存或重用,
// 该对象生命期总是短于 Impl. 该对象线程不安全
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
  Bind(name string, data interface{}) error
  //
  // 关闭这个上下文, 实现通常会关闭底层网络连接, 重复调用该方法是安全的.
  //
  Close() error
  //
  // 上下文已经关闭返回 true
  //
  Closed() bool
  //
  // 对于该函数的使用方, DTU 中的数据转换是透明的, 
  // 用户只关心与设备之间的原始数据格式.
  // 如果 dtu 不支持该方法则总是返回错误,
  // Conn.Close() 方法与 Context.Close() 有相同的效果
  //
  GetConn() (net.Conn, error)
}


//
// 接受 dtu 消息, 需要保证线程安全
//
type Event interface {
  //
  // 创建新的上下文, 通常是因为一个新的客户端发起了请求.
  // 如果创建失败则 ctx 为空, e 将保存失败原因
  //
  NewContext(ctx Context, e error)
  //
  // 在启动成功后, 任何关闭了 DTU 的行为都会调用该方法.
  // 如果是因为错误导致 DTU 终止, 则带有 error 参数, 否则为空.
  //
  OnClose(error)
  //
  // DTU 启动消息
  //
  OnStart(msg string)
}