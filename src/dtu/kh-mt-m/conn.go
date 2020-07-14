package kh_mt_m

import (
	"errors"
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
      return -1, errors.New("远程连接已关闭 "+ c.RemoteAddr().String())
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
    log.Printf("%s %t", se.Unwrap().Error(), se.Unwrap())
    c.Close()
    log.Println("close on write")
  }
  return
}