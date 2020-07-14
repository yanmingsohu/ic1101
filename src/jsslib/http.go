package jsslib

import (
	"bytes"
	"encoding/json"
	"errors"
	"ic1101/src/core"
	"ic1101/src/js"
	"io"
	"log"
	"net/http"

	"github.com/dop251/goja"
)

const MAX_BODY_LEN = 3 * core.MB
const BODY_TYPE = "application/octet-stream"


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
  return h.response(res)
}


func (h *JSHttp) Post(f goja.FunctionCall) goja.Value {
  url := f.Argument(0).String()
  if url == "" {
    panic(errors.New("URL 参数不能为空"))
  }
  body := f.Argument(1).Export()
  if body == nil {
    panic(errors.New("Body 参数不能为空"))
  }

  by, err := json.Marshal(body)
  if err != nil {
    panic(err)
  }
  reader := bytes.NewReader(by)

  res, err := http.Post(url, BODY_TYPE, reader)
  if err != nil {
    panic(err)
  }
  return h.response(res)
}


func (h *JSHttp) response(res *http.Response) *goja.Object {
  len := res.ContentLength
  if len > MAX_BODY_LEN {
    len = MAX_BODY_LEN
  } else if len < 0 {
    len = 1024
  }

  ret := h.NewObject()
  buf := make([]byte, len)
  n, _ := io.ReadFull(res.Body, buf[:])

  ret.Set("status", res.StatusCode)
  ret.Set("body",   buf[:n])
  ret.Set("header", res.Header)
  return ret
}