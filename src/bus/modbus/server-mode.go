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
package bus_modbus

import (
	"errors"
	"ic1101/src/bus"
	"ic1101/src/dtu"
	"net/url"
	"strconv"
	"time"

	"github.com/simonvetter/modbus"
)

const MBCLIENT = "modbus_client"

const (
  MODE_TCP = 1
  MODE_RTU = 2
)


type tcp_server struct {
  dtu    dtu.Impl
  rel    bus.BusReal
  parm   UrlParam
  nr     bus.NotRepeating
}


func (r *tcp_server) Start(i bus.BusReal) (error) {
  d, err := dtu.GetDtu(i.URL(), r)
  if err != nil {
    return err
  }
  r.dtu = d
  r.rel = i
  r.parm = UrlParam{}
  r.parm.Parse(i.URL())
  r.nr = bus.NewNotRep(i)
  return nil
}


func (r *tcp_server) Stop(i bus.BusReal) {
  r.dtu.Stop()
  i.Log("总线停止")
}


func (r *tcp_server) get_client(id int) (*MC, error) {
  ctx, err := r.dtu.GetContext(id)
  if err != nil {
    return nil, err
  }
  tmp, err := ctx.Get(MBCLIENT)
  if err != nil {
    return nil, err
  }
  return tmp.(*MC), nil
}


func (r *tcp_server) SyncData(i bus.BusReal, t *time.Time) error {
  for _, s := range i.Datas() {
    ms := s.(*modbus_slot)
    
    client, err := r.get_client(int(ms.c))
    if err != nil {
      r.nr.Log(ms.LogicAddr(), ms.ErrInfo(err))
      continue
    }
    
    if r.parm.sid >= 0 {
      client.SetUnitId(uint8(r.parm.sid))
    } else {
      client.SetUnitId(ms.c)
    }
    client.setMode(ms.l)

    d, err := client.read(ms)
    if err != nil {
      i.Log(err.Error())
      continue
    }

    i.Event().OnData(s, t, d)
    r.nr.Recover(ms.LogicAddr(), ms.ErrInfo("已恢复"))
  }
  return nil
}


func (r *tcp_server) SendCtrl(_s bus.Slot, d bus.DataWrap, t *time.Time) error {
  s := _s.(*modbus_slot)
  client, err := r.get_client(int(s.c))
  if err != nil {
    if r.nr.Log(s.LogicAddr(), s.ErrInfo("发送控制")) {
      return errors.New(s.ErrInfo(err))
    }
    return nil
  }
  
  if r.parm.sid >= 0 {
    client.SetUnitId(uint8(r.parm.sid))
  } else {
    client.SetUnitId(s.c)
  }
  client.setMode(s.l)
  r.nr.Recover(s.LogicAddr(), s.ErrInfo("已恢复"))
  return client.send(s, d)
}


func (r *tcp_server) NewContext(ctx dtu.Context, err error) {
  if err != nil {
    r.rel.Log("远程连接失败, "+ err.Error())
    return
  }

  var url string
  switch r.parm.mode {
  case MODE_RTU:
    url = "rtuovertcp://localhost"
  case MODE_TCP:
    url = "tcp://localhost"
  default:
    url = "tcp://localhost"
  }

  client, err := modbus.NewClient(&modbus.ClientConfiguration{
    URL:      url,
    Timeout:  r.parm.timeout,
  })

  if err != nil {
    r.rel.Log("创建 modbus 连接失败 "+ err.Error())
    ctx.Close()
    return;
  }

  conn, err := ctx.GetConn()
  if err != nil {
    r.rel.Log("获取客户端连接失败 "+ err.Error())
    ctx.Close()
    return;
  }

  if err = client.Bind(conn); err != nil {
    r.rel.Log("绑定 Socket 错误 "+ err.Error())
    ctx.Close()
    return;
  }

  if err = ctx.Bind(MBCLIENT, &MC{client}); err != nil {
    r.rel.Log("绑定参数错误 "+ err.Error())
    ctx.Close()
    return;
  }
  
  r.rel.Log("远程已经连接", conn.RemoteAddr(), "从机地址", ctx.Id())
}


func (r *tcp_server) OnClose(e error) {
  if e != nil {
    r.rel.Log("DTU 因错误终止 "+ e.Error())
  }
}


func (r *tcp_server) OnStart(msg string) {
  r.rel.Log("总线启动, " + msg)
}


type UrlParam struct {
  mode      uint
  timeout   time.Duration
  sid       int
}


func (p *UrlParam) Parse(u *url.URL) {
  vs := u.Query()
  // mode = {rtu|tcp}
  m := vs.Get("mode")
  if m == "rtu" {
    p.mode = MODE_RTU
  } else {
    p.mode = MODE_TCP
  }

  // timeout = 1 ~ MAX
  t := vs.Get("timeout")
  p.timeout = 10 * time.Second 
  if t != "" {
    i, _ := strconv.Atoi(t)
    if i > 0 {
      p.timeout = time.Duration(i) * time.Second
    }
  }

  // sid < 0 使用动态从机地址, 否则固定从机地址
  s := vs.Get("sid")
  p.sid = -1
  if s != "" {
    i, _ := strconv.Atoi(s)
    if i >= 0 {
      p.sid = i
    }
  }
}