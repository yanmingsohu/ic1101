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
	"ic1101/src/dtu"
	"io"
	"log"
	"net"
)


type conn_wrap struct {
  dtu.ConnWrap
  rd *dtu.RemoveDirty
}


//
// 该方法并不保证读取多少字节
//
func (c *conn_wrap) Read(buf []byte) (int, error) {
  end, err := c.Conn.Read(buf)
  if err != nil {
    if err == io.EOF {
      c.Close()
      // return -1, errors.New("远程连接已关闭 "+ c.RemoteAddr().String())
    }
    return -1, err
  }
  end = c.rd.Modify(buf[:end])

  log.Printf("R %s %3d < % x", c.RemoteAddr().String(), end, buf[:end])
  return end, err
}


func (c *conn_wrap) Write(b []byte) (n int, err error) {
  log.Printf("W %s %3d > % x", c.RemoteAddr().String(), len(b), b[:])

  n, err = c.Conn.Write(b)
  if se, ok := err.(*net.OpError); ok && (se.Temporary()==false) {
    c.Close()
    // log.Println("Close on write", se.Unwrap().Error())
  }
  return
}