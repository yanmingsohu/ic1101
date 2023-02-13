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
	"bytes"
	"net"
	"time"
)


//
// 这是 net.Conn 的包装器, 除了 Read/Write 方法之外都直接调用 Conn 的方法.
//
type ConnWrap struct {
  Conn net.Conn
  ctx  Context
}


func (c *ConnWrap) InitWrap(ctx Context, conn net.Conn) {
  c.Conn = conn
  c.ctx = ctx
}


func (c *ConnWrap) Close() error {
  return c.ctx.Close()
}


func (c *ConnWrap) LocalAddr() net.Addr {
  return c.Conn.LocalAddr()
}


func (c *ConnWrap) RemoteAddr() net.Addr {
  return c.Conn.RemoteAddr()
}


func (c *ConnWrap) SetDeadline(t time.Time) error {
  return c.Conn.SetDeadline(t)
}


func (c *ConnWrap) SetReadDeadline(t time.Time) error {
  return c.Conn.SetReadDeadline(t)
}


func (c *ConnWrap) SetWriteDeadline(t time.Time) error {
  return c.Conn.SetWriteDeadline(t)
}


//
// 部分 dtu 的协议设计, 总是定时发送一个脏数据用于心跳包
// 该对象记录脏数据的格式, 并从数据流中删除这些脏数据.
//
type RemoveDirty struct {
  state int
  dirty []byte
  dl    int
}


func NewRemoveDirty(dirty []byte) *RemoveDirty {
  return &RemoveDirty{0, dirty, len(dirty)}
}


//
// 从 has_dirty 中删除 drity 数据并返回最终数据长度
//
func (r *RemoveDirty) Modify(has_dirty []byte) int {
  end := len(has_dirty)
  begin := 0
  for {
    if i := bytes.Index(has_dirty[begin:], r.dirty); i>=0 {
      i += begin
      j := i + r.dl 
      if j < end {
        for x := j; x < end; x++ {
          has_dirty[x-r.dl] = has_dirty[x]
          has_dirty[x] = 0
        }
      } else {
        end -= r.dl
        break;
      }
      end -= r.dl
      begin = i + r.dl
    } else {
      break;
    }
  }
  return end
}