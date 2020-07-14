package kh_mt_m

import (
	"errors"
	"ic1101/src/dtu"
	"io"
	"log"
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
  log.Printf("Read %s %d - [% x]", c.RemoteAddr().String(), end, buf[:end])
  return end, err
}


func (c *conn_wrap) Write(b []byte) (n int, err error) {
  log.Printf("Write %s - [% x]", c.RemoteAddr().String(), b[:])
  return c.Conn.Write(b)
}