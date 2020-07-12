package service

import (
	"context"
	"ic1101/brick"
	"ic1101/src/bus"
	"ic1101/src/core"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/dop251/goja"
)

const NullScript = `
{
  //TODO: 系统生成空脚本
  on_data : function(dev, time, data) {
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
  sr.Name = name
  return &sr, nil
}


//
// 设备通过该类型导出的方法, 将数据转换后保存到 DB
// 线程不安全
//
type ScriptRuntime struct {
  core.ScriptRuntime
  on_data   goja.Callable
  Name      string
}


//
// 调用脚本导出的 on_data(dev, time, data) 函数
//
func (s *ScriptRuntime) OnData(slot *core.BusSlot, 
    t *time.Time, d bus.DataWrap) (bus.DataWrap, error) {
  jsdev  := JSDevData{slot, s}
  jstime := s.Value(t.UnixNano() / 1e6)
  jsdata := s.Value(d.Src())
  // js: Function(dev, timeMS, data)
  ret, err := s.on_data(s.This(), s.Value(&jsdev), jstime, jsdata)
  if err != nil {
    return d, err
  }
  return bus.NewDataWrap(ret.Export())
}


//
// 该对象导出到 js 环境中, 作为 on_data 方法的参数
//
type JSDevData struct {
  data    *core.BusSlot
  sr      *ScriptRuntime
}


//
// 返回数据名
// js: String GetName()
//
func (d *JSDevData) GetName(fc goja.FunctionCall) goja.Value {
  return d.sr.Value(d.data.Name)
}


//
// 返回数据槽 id 地址, 这个 id 在不同的协议上使用不同的格式
// js: String GetSlot()
//
func (d *JSDevData) GetSlot(fc goja.FunctionCall) goja.Value {
  return d.sr.Value(d.data.SlotID)
}


//
// 返回设备 id
// js: String GetDev()
//
func (d *JSDevData) GetDev(fc goja.FunctionCall) goja.Value {
  return d.sr.Value(d.data.Dev)
}


//
// 返回数据类型
// js: Int GetType()
//
func (d *JSDevData) GetType(fc goja.FunctionCall) goja.Value {
  return d.sr.Value(d.data.Type)
}


//
// 读取脚本
//
func GetDevScript(ctx context.Context, script_id string, ret interface{}) (error) {
  err := mg.GetOne(core.TableDevScript, ctx, script_id, ret)
  if err != nil {
    return err
  }
  return nil
}