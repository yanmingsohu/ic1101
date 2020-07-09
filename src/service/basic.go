package service

import (
	"fmt"
	"log"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"ic1101/brick"
	"ic1101/src/core"

	"go.mongodb.org/mongo-driver/mongo"
)

var root = core.LoginUser{}
var mg *core.Mongo
var salt string
var PageSize int64 = 10
var auth_arr = []string{}
var sessionExp = 20 * time.Hour


func serviceList(b *brick.Brick) {
  installUserService(b)
  installDictService(b)
  installAuthService(b)
  installDevProtoService(b)
  installDeviceService(b)
  installTimerService(b)
  installBusService(b)
  installLogService(b)
  installDeviceDataService(b)
  installLicenseService(b)
}


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


//
// 服务接口参数包装器
//
type Ht struct {
  *brick.Http
  *ServiceGroupContext
}


//
// 服务组上下文, 服务组中的所有服务都使用相同的配置
//
type ServiceGroupContext struct {
  // 操作数据表的名字
  collectionName string
  // 服务的描述名
  serviceName    string
}


//
// 服务接口签名, 返回 error 则返回错误 json, 
// 返回 HttpRet 则返回该对象的 json, 
// 返回其他数据则绑定到 json.data 属性上.
//
type ServiceHandler func(*Ht) interface{}


func Install(conf *core.Config, m *core.Mongo) {
  mg = m;
  salt = conf.Salt
  core.InitRootUser(&root)
  
  root.Pass = encPass(root.Name, root.Pass)
  root.Pass = encPass(root.Name, root.Pass)

  b := brick.NewBrick(conf.HttpPort, sessionExp)
  b.SetErrorHandler(httpErrorHandle)

  b.StaticPage("/ic/ui", "www")
  b.Service("/ic/", notfound)
  serviceList(b)

  b.HttpJumpMapping("/", "/ic/ui/index.html")
  b.HttpJumpMapping("/favicon.ico", "/ic/ui/favicon.ico")
  b.StartHttpServer()
}


func notfound(h brick.Http) error {
  h.W.WriteHeader(404)
  h.Json(HttpRet{404, "Api Not found", h.R.URL.Path})
  return nil
}


//
// 无授权检测
//
func dserv(b *brick.Brick, ctx *ServiceGroupContext, 
           name string, service ServiceHandler) {
  // name := funcName(h)
  b.Service("/ic/"+ name, func(h brick.Http) error {
    ht := Ht{&h, ctx}
    ret := service(&ht)
    
    if ret != nil {
      switch ret.(type) { 
      case error:
        return ret.(error)
        
      case HttpRet:
        ht.Json(ret.(HttpRet))
        break;

      default:
        ht.Json(HttpRet{0, "data", ret})
        break;
      }
    }
    
    return nil
  })
}


//
// 检查登录/权限/ TODO:软件授权
//
func aserv(b *brick.Brick, ctx *ServiceGroupContext, 
           name string, handler ServiceHandler) {
  auth_arr = append(auth_arr, name)

  dserv(b, ctx, name, func(h *Ht) interface{} {
    v := h.Session().Get("user")
    if v == nil {
      h.Json(HttpRet{100, "用户未登录", nil})
      return nil
    }

    user := v.(*core.LoginUser)
    if !user.IsRoot {
      log.Print("[", user.Name, ":", name, "] No auth")
      if !user.Auths[name] {
        h.Json(HttpRet{101, "用户无权限操作", nil})
        return nil
      }
    } else {
      log.Print("[", user.Name, ":", name, "] ", h.Get("id"))
    }

    return handler(h)
  })
}


//
// 必须是超级管理员用户, 没有其他限制
//
func lserv(b *brick.Brick, ctx *ServiceGroupContext, 
           name string, handler ServiceHandler) {
  dserv(b, ctx, name, func(h *Ht) interface{} {
    v := h.Session().Get("user")
    if v == nil {
      h.Json(HttpRet{100, "用户未登录", nil})
      return nil
    }
    user := v.(*core.LoginUser)
    if !user.IsRoot {
      h.Json(HttpRet{101, "只有超级用户可以执行该操作", nil})
    }
    log.Print("[", user.Name, ":", name, "] ", h.Get("id"))

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
    msg := fmt.Sprintf("%s %s %d %s", info, "长度必须大于等于", min, "个字符")
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


func checkpage(h *Ht) int64 {
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


func checkint(info string, s string, min int64, max int64) int64 {
  r, err := strconv.ParseInt(s, 10, 32) 
  if err != nil {
    r = 0
  }
  if r < min {
    msg := fmt.Sprintf("%s %s %d", info, "必须大于等于", min)
    panic(HttpRet{ Code: 2, Msg: msg, Data: []int64{min, max} })
  }
  if r >= max {
    msg := fmt.Sprintf("%s %s %d", info, "必须小于", max)
    panic(HttpRet{ Code: 2, Msg: msg, Data: []int64{min, max} })
  }
  return r
}



func (h *HttpRet) Error() string {
  return h.Msg.(string)
}


//
// 返回当前登录用户
//
func (h *Ht) GetUser() *core.LoginUser {
  return h.Session().Get("user").(*core.LoginUser)
}


//
// 返回 CRUD 实例, 配置的 db 表由绑定服务接口时的 ServiceGroupContext 参数决定.
//
func (h *Ht) Crud() *Crud {
  return &Crud{h, h.collectionName, h.serviceName}
}


func (h *Ht) Table() *mongo.Collection {
  return mg.Collection(h.collectionName)
}


//
// 如果参数不存在或解析错误返回 defaultVal, 否则返回参数的 int 值
//
func (h *Ht) GetInt(name string, defaultVal int) int {
  v := h.Get(name)
  if "" == v {
    return defaultVal
  }
  i, err := strconv.Atoi(v) 
  if err != nil {
    return defaultVal
  }
  return i
}


//
// 返回 bool 值, 不会抛出任何异常
//
func (h *Ht) GetBool(name string) bool {
  return checkbool(name, h.Get(name))
}