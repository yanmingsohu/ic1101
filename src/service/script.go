package service

import (
	"context"
	"ic1101/brick"
	"ic1101/src/core"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/dop251/goja"
)

const NullScript = `
{
  //
  // time: 数据时间
  // data: 数据包装器
  // 完整示例见开发文档
  //
  on_data : function(time, data) {
    // 默认直接返回原始值
    return data;
  },
}
`

func installScriptService(b *brick.Brick) {
  mg.CreateIndex(core.TableDevScript, &bson.D{
    {"_id", "text"}, {"desc", "text"}, {"js", "text"} })
  ctx := &ServiceGroupContext{core.TableDevScript, "设备脚本"}
  
  aserv(b, ctx, "dev_sc_count",   dev_sc_count)
  aserv(b, ctx, "dev_sc_list",    dev_sc_list)
  aserv(b, ctx, "dev_sc_read",    dev_sc_read)
  aserv(b, ctx, "dev_sc_delete",  dev_sc_delete)
  aserv(b, ctx, "dev_sc_create",  dev_sc_create)
  aserv(b, ctx, "dev_sc_update",  dev_sc_update)
}


func dev_sc_count(h *Ht) interface{} {
  return h.Crud().PageInfo()
}


func dev_sc_list(h *Ht) interface{} {
  return h.Crud().List(func(opt *options.FindOptions) {
    opt.SetProjection(bson.M{ "js":0 })
  }) 
}


func dev_sc_read(h *Ht) interface{} {
  id := checkstring("脚本ID", h.Get("id"), 2, 20)
  return h.Crud().Read(id)
}


func dev_sc_delete(h *Ht) interface{} {
  id := checkstring("脚本ID", h.Get("id"), 2, 20)
  return h.Crud().Delete(id)
}


func dev_sc_create(h *Ht) interface{} {
  d := bson.M{
    "_id"     : checkstring("脚本ID", h.Get("id"), 2, 20),
    "desc"    : checkstring("脚本说明", h.Get("desc"), 0, 99),
    "cd"      : time.Now(),
    "md"      : "",
    "js"      : NullScript,
    "size"    : len(NullScript),
    "version" : 1,
  }

  return h.Crud().Create(&d)
}


func dev_sc_update(h *Ht) interface{} {
  if err := h.R.ParseForm(); err != nil {
    return err
  }
  qr := h.R.PostForm
  
  id := checkstring("脚本ID", qr.Get("id"), 2, 20)
  js := checkstring("js脚本", qr.Get("js"), 1, 9999999)
  sr := core.ScriptRuntime{}
  if err := sr.Compile(id, js); err != nil {
    return HttpRet{1, "编译失败", err.Error()}
  }

  up := bson.M{
    "$inc" : bson.M{ "version" : 1 },
    "$set" : bson.M{
      "desc"   : qr.Get("desc"), 
      "md"     : time.Now(),
      "js"     : js,
      "size"   : len(js),
    },
  }
  return h.Crud().Update(id, up)
}


type ScriptRuntime struct {
  core.ScriptRuntime
  on_data   goja.Callable
}


func BuildDevScript(name, code string) (*ScriptRuntime, error) {
  sr := ScriptRuntime{}
  if err := sr.Compile(name, code); err != nil {
    return nil, err
  }
  if err := sr.InitObject(); err != nil {
    return nil, err
  }
  on_data, err := sr.GetFunc("on_data")
  if err != nil {
    return nil, err
  }
  sr.on_data = on_data
  return &sr, nil
}


//
// 读取脚本
//
func GetDevScript(script_id string, ctx context.Context) (*core.DevScript, error) {
  d := core.DevScript{}
  err := mg.GetOne(core.TableDevScript, ctx, script_id, &d)
  if err != nil {
    return nil, err
  }
  return &d, nil
}