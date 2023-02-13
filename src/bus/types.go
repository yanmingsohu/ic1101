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
	"net/url"
	"time"
)


//
// 总线运行状态
//
type BusState int

const (
  BusStateStop BusState = iota       
  BusStateStartup
  BusStateSleep
  BusStateTask
  BusStateShutdown  BusState = -1 
  BusStateFailStart BusState = -2
)

func (s BusState) String() string {
  switch s {
  case BusStateStop:
    return "停止"
  case BusStateStartup:
    return "正在启动"
  case BusStateSleep:
    return "休眠中, 等待计时器"
  case BusStateTask:
    return "执行任务"
  case BusStateShutdown:
    return "关闭中"
  case BusStateFailStart:
    return "启动失败"
  }
  return "无效"
}

//
// 槽的类型
//
type SlotType int

const (
  SlotInvaild SlotType = iota
  SlotData    SlotType = iota
  SlotCtrl    SlotType = iota
)

func (t SlotType) String() string {
  switch (t) {
  case SlotData:
    return "数据槽"
  case SlotCtrl:
    return "控制槽"
  }
  return "无效"
}


//
// 总线创建接口
//
type BusCreator interface {
  SlotParser
  // 创建总线实例
  Create(BusReal) (Bus, error)
  // 总线的显示名称
  Name() string
}


//
// 运行中的总线实例对象, 该对象的操作已经被包装
// 无需考虑线程安全问题.
//
type Bus interface {
  // 启动总线, 用于初始化数据, 失败返回 error
  // 在该方法中失败, 不会调用 Stop
  Start(BusReal) error
  // 传送一次数据
  SyncData(BusReal, *time.Time) error
  // 停止总线, 该方法返回后, 总线一定是停止的
  Stop(BusReal)
  // 发送控制指令, 发送失败返回 error
  Sender
}


//
// 发送命令的接口
//
type Sender interface {
  // 发送控制指令, 发送失败返回 error
  SendCtrl(s Slot, d DataWrap, t *time.Time) error
}


type Logger interface {
  // 记录日志
  Log(msg ...interface{})
}


//
// 该对象是面向总线实现者的, 尽可能暴露方法
//
type BusReal interface {
  Logger
  // 返回创建总线使用的 url
  URL() *url.URL
  // 返回事件对象
  Event() BusEvent
  // 返回数据槽的切片
  Datas() []Slot
}


//
// 数据包装器接口
//
type DataWrap interface {
  Int()     int
  Int64()   int64
  String()  string
  Float()   float32
  Float64() float64
  Bool()    bool
  // 返回内部类型的值
  Src() interface{}
}


//
// 消息接收器接口
//
type BusEvent interface {
  // 接受总线发送的数据, [该方法由总线调用, 参数一定不为 nil]
  OnData(slot Slot, time *time.Time, data DataWrap)
  // 总线上所有任务都终止了, 该方法被调用
  OnStopped()
  // 一个控制命令已经发出后该方法被调用
  OnCtrlSended(s Slot, t *time.Time)
  // 一个控制槽已经终止任务该方法被调用
  OnCtrlExit(s Slot)
}


//
// 总线数据/控制槽
//
type Slot interface {
  // 返回插槽的唯一标识名称, 该标识是对 slot 描述的完整序列化
  // 与 BusCreator 中的 ParseSlot 对应.
  // 总线实例需要判断 '数据/控制' 类型
  String() string
  // 返回对插槽的可读文本
  Desc() string
  Type() SlotType
}


//
// 用于解析 slot, 数据槽格式包含 '数据/控制' 的描述
//
type SlotParser interface {
  // 通过字符串 (序列化的slot) 解析 slot 实例, 失败返回 error
  // 每种总线都有自己的 slot 格式
  ParseSlot(s string) (Slot, error)
  // 返回可读的对端口的描述字符串, 格式无效返回 error
  SlotDesc(s string) (string, error)
  // 解析 uri 用于创建服务器, 如果 uri 无效, 或不受支持返回 error
  ParseURI(uri string) (*url.URL, error)
}