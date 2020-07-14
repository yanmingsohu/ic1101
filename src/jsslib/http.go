package jsslib

import (
	"errors"
	"ic1101/src/core"
	"ic1101/src/js"
	"io"
	"log"
	"net/http"

	"github.com/dop251/goja"
)

const MAX_BODY_LEN = 3 * core.MB


func init() {
  js.Reg("http", &http_fact{})
}


type http_fact struct {}


func (*http_fact) New(v js.JSValue) interface{} {
  return &JSHttp{v}
}


type JSHttp struct {
  js.JSValue
}


func (h *JSHttp) Send(f goja.FunctionCall) goja.Value {
  url := f.Argument(0).String()
  if url == "" {
    panic(errors.New("URL 参数不能为空"))
  }
  go (func() {
    _, err := http.Get(url)
    if err != nil {
      log.Println("http.send", err)
      return
    }
  })()
  return h.Value(nil)
}


func (h *JSHttp) Get(f goja.FunctionCall) goja.Value {
  url := f.Argument(0).String()
  if url == "" {
    panic(errors.New("URL 参数不能为空"))
  }
  res, err := http.Get(url)
  if err != nil {
    panic(err)
  }
  len := res.ContentLength
  if len > MAX_BODY_LEN {
    len = MAX_BODY_LEN
  }

  ret := h.NewObject()
  buf := make([]byte, len)
  io.ReadFull(res.Body, buf)

  ret.Set("status", res.Status)
  ret.Set("body",   buf)
  ret.Set("header", res.Header)
  return ret
}


func (h *JSHttp) Post(f goja.FunctionCall) goja.Value {
  return h.Value(nil)
}