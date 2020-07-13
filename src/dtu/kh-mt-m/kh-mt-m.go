//
// <北京科慧铭远自控技术有限公司> 生产的 dtu
// http://www.msi-automation.com/
//
package kh_mt_m

import (
	"errors"
	"ic1101/src/dtu"
	"io"
	"log"
	"net"
	"net/url"
	"time"
)


var HeartBeat = []byte("KHMY")
var timeout = 10 * time.Second


func init() {
  dtu.RegFact(&factory{})
}


type factory struct {
}


func (f *factory) Name() string {
  return "kh-mt-m"
}


func (f *factory) Desc() string {
  return "科慧铭远 MBUS-MODBUS-DTU"
}


func (f *factory) New(url *url.URL, handle dtu.Event) (dtu.Impl, error) {
  d := dtu_impl{
    url     : url, 
    event   : handle, 
  }
  if err := d.init(); err != nil {
    return nil, err
  }
  return &d, nil
}


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
  // d.Lock()
  // defer d.Unlock()
  if !d.Run {
    return 
  }

  d.Run = false
  d.CtxMap = nil
  d.serv.Close()
  d.event.OnClose(err)
}


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
      return -1, errors.New("远程连接已关闭 "+ c.RemoteAddr().String())
    }
    return -1, err
  }
  end = c.rd.Modify(buf[:end])
  log.Printf("Read %s %d - [% x]", c.RemoteAddr().String(), end, buf[:end])
  return end, err
}


func (c *conn_wrap) Write(b []byte) (n int, err error) {
  log.Printf("Write %s - [% x]", c.RemoteAddr().String(), b[:])
  return c.Conn.Write(b)
}