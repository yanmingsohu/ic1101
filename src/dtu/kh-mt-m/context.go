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
package kh_mt_m

import (
	"errors"
	"ic1101/src/dtu"
	"io"
	"log"
	"net"
	"time"
)

var HeartBeat = []byte("KHMY")
var timeout = 10 * time.Second


type ctx struct {
  dtu.CtxHelp

  conn net.Conn
  cw   *conn_wrap
}


// 
// DTU 连接后立即发送一个帧头, 帧头检测错误直接关闭连接:
// FE  帧头
// 00  ID高字节
// 01  ID低字节
// 01  校验位-求和校验（00 + 01）
// FE  帧尾
//
// 之后 DTU 定时发送 4 个字节心跳包:
// KHMY
//
func (c *ctx) init(conn net.Conn, di *dtu_impl) (err error) {
  conn.SetDeadline(time.Now().Add(timeout))

  b := make([]byte, 5)
  n, err := io.ReadFull(conn, b)
  if err != nil {
    return
  }

  log.Printf("接受握手帧 %s [% x]", conn.RemoteAddr(), b)
  if n != 5 {
    err = errors.New("连接已关闭, 无效的首帧长度")
  }
  if b[0] != 0xFE || b[4] != 0xFE {
    err = errors.New("连接已关闭, 无效的帧头")
  }
  if b[3] != b[1] + b[2] {
    err = errors.New("连接已关闭, 无效的帧头校验")
  }
  if err != nil {
    conn.Close()
    return
  }

  id := int(uint(b[1]) << 8) + int(b[2])
  c.InitHelp(id, di)
  c.conn = conn
  c.cw = &conn_wrap{ rd : dtu.NewRemoveDirty(HeartBeat) }
  c.cw.InitWrap(c, conn)
  return nil
}


func (c *ctx) Close() error {
  if (c.CloseHelp()) {
    return c.conn.Close()
  }
  return nil
}


func (c *ctx) GetConn() (net.Conn, error) {
  return c.cw, nil
}
