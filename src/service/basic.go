package service

import (
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"ic1101/brick"
	"ic1101/src/core"
)

var root = core.LoginUser{}
var mg *core.Mongo
var salt string
var PageSize int64 = 10


//
// Code 定义:
//    0: 无错误
//    1: 一般错误
//    2: 参数错误
//  100: 无权操作
//
type HttpRet struct {
  Code int					`json:"code"`
  Msg  interface{}	`json:"msg"`
  Data interface{}	`json:"data"`
}


func (h *HttpRet) Error() string {
  return h.Msg.(string)
}


func Install(conf *core.Config, m *core.Mongo) {
  mg = m;
  salt = conf.Salt
  core.InitRootUser(&root)
  
  root.Pass = encPass(root.Name, root.Pass)
  root.Pass = encPass(root.Name, root.Pass)

  b := brick.NewBrick(conf.HttpPort)
  b.SetErrorHandler(httpErrorHandle)

  b.StaticPage("/ic/ui", "www")
  serviceList(b)

  b.HttpJumpMapping("/", "/ic/ui/index.html")
  b.StartHttpServer()
}


func serviceList(b *brick.Brick) {
  installUserService(b)
  installDictService(b)
}


// 无授权检测
func dserv(b *brick.Brick, name string, h brick.HttpHandler) {
  // name := funcName(h)
  b.Service("/ic/"+ name, h)
}


// 检查登录/权限
func aserv(b *brick.Brick, name string, handler brick.HttpHandler) {
  b.Service("/ic/"+ name, func(h brick.Http) error {
    v := h.Session().Get("user")
    if v == nil {
      h.Json(HttpRet{100, "用户未登录", nil})
      return nil
    }
    return handler(h)
  })
}


func funcName(h interface{}) string {
  name := runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
  i := strings.Index(name, ".")
  if i >= 0 {
    return name[i+1:]
  }
  return name
}


func httpErrorHandle(hd *brick.Http, err interface{}) {
  // log.Print("Error:", err)
  var msg string

  switch err.(type) {

  case HttpRet:
    hd.Json(err.(HttpRet))
    return

  case error:
    msg = err.(error).Error()
    break

  case string:
    msg = err.(string)
    break
  }
  
  ret := HttpRet{ Code: 1, Msg: msg, Data: err }
  hd.Json(ret)
}


//
// 如果字符串 s 超出 min,max 指定范围抛出异常
// 
func checkstring(info string, s string, min int, max int) string {
  l := len(s)
  if l < min {
    msg := fmt.Sprintf("%s %s %d %s", info, "长度必须大于", min, "个字符")
    panic(HttpRet{ Code: 2, Msg: msg, Data: []int{min, max} })
  }
  if l >= max {
    msg := fmt.Sprintf("%s %s %d %s", info, "长度必须小于", max, "个字符")
    panic(HttpRet{ Code: 2, Msg: msg, Data: []int{min, max} })
  }
  return s
}


func checkbool(info string, v string) bool {
  if v == "on" || v == "ON" {
    return true
  }
  b, _ := strconv.ParseBool(v)
  return b
}


func checkpage(h brick.Http) int64 {
  parm := h.Get("page")
  if parm == "" {
    return 0
  }
  page, err := strconv.ParseInt(parm, 10, 32) 
  if err != nil {
    return 0
  }
  return page * PageSize
}