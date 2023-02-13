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
//
// <北京科慧铭远自控技术有限公司> 生产的 dtu
// http://www.msi-automation.com/
//
package kh_mt_m

import (
	"ic1101/src/dtu"
	"net"
	"net/url"
)


type dtu_impl struct {
  dtu.ImplHelp
  
  url     *url.URL
  event   dtu.Event
  serv    net.Listener
}


func (d *dtu_impl) init() error {
  d.ImplHelp = dtu.NewImplHelp()
  serv, err := net.Listen("tcp", d.url.Host)
  if err != nil {
    return err
  }
  d.serv = serv

  go d.accept()
  return nil
}


func (d *dtu_impl) accept() {
  d.event.OnStart(d.serv.Addr().String())
  
  for d.Run {
    conn, err := d.serv.Accept()
    if err != nil {
      d.stop(err)
    } else {
      d.new_context(conn)
    }
  }
}


func (d *dtu_impl) new_context(conn net.Conn) {
  c := ctx{}
  if err := c.init(conn, d); err != nil {
    d.event.NewContext(nil, err)
  } else {
    if err := d.SaveContext(&c); err != nil {
      d.event.NewContext(nil, err)
    } else {
      d.event.NewContext(&c, nil)
    }
  }
}


func (d *dtu_impl) Stop() {
  d.stop(nil)
}


func (d *dtu_impl) stop(err error) {
  if !d.Run {
    return 
  }

  d.Run = false
  d.CtxMap = nil
  d.serv.Close()
  d.event.OnClose(err)
}
