package service

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"../../brick"
	"../core"
)

var root = core.LoginUser{}
var mg *core.Mongo
var salt string


type HttpRet struct {
	Code int					`json:"code"`
	Msg  interface{}	`json:"msg"`
	Data interface{}	`json:"data"`
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
	dserv(b, "login", 			login)
	dserv(b, "logout",  		logout)
	dserv(b, "salt",  			getsalt)

	aserv(b, "whoaim",  		whoaim)
	aserv(b, "reguser", 		reguser)
	aserv(b, "changepass", 	changepass)
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
			return errors.New("用户未登录")
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
	case error:
		msg = err.(error).Error()

	case string:
		msg = err.(string)

	case HttpRet:
		hd.Json(err.(HttpRet))
		return;
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
		panic(fmt.Sprintf("%s %s %d %s", info, "长度必须大于", min, "个字符"))
	}
	if l >= max {
		panic(fmt.Sprintf("%s %s %d %s", info, "长度必须小于", max, "个字符"))
	}
	return s
}