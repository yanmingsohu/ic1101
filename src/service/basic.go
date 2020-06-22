package service

import (
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


type User struct {
	Name string
	Auth interface{}
}


func Install(conf *core.Config, m *core.Mongo) {
	mg = m;
	salt = core.RandStringRunes(20)
	core.InitRootUser(&root)
	
	root.Pass = encPass(root.Name, root.Pass)
	root.Pass = encPass(root.Name, root.Pass)

	b := brick.NewBrick(conf.HttpPort)
	b.SetErrorHandler(httpErrorHandle)

	b.StaticPage("/ic/ui", "www")
	
	dserv(b, "login", login)
	dserv(b, "salt",  getsalt)

	b.HttpJumpMapping("/", "/ic/ui/index.html")
	b.StartHttpServer()
}


// 无授权检测
func dserv(b *brick.Brick, name string, h brick.HttpHandler) {
	// name := funcName(h)
	b.Service("/ic/"+ name, h)
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
	ret := HttpRet{ Code: 1, Msg: err.(error).Error() }
	hd.Json(ret)
}