package service

import (
	"context"
	"errors"
	"ic1101/brick"
	"ic1101/src/core"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const INT_MAX = int64(^uint64(0) >> 1)
const INT_MIN = ^INT_MAX


func installDevProtoService(b *brick.Brick) {
  mg.CreateIndex(core.TableDevProto, &bson.D{{"_id", "text"}, {"desc", "text"}})
  ctx := &ServiceGroupContext{core.TableDevProto, "设备原型"}

  aserv(b, ctx, "dev_proto_create",       dev_proto_create)
  aserv(b, ctx, "dev_proto_count",        dev_proto_count)
  aserv(b, ctx, "dev_proto_list",         dev_proto_list)
  aserv(b, ctx, "dev_proto_update",       dev_proto_update)
  aserv(b, ctx, "dev_proto_delete",       dev_proto_delete)
  aserv(b, ctx, "dev_proto_read",         dev_proto_read)

  dserv(b, ctx, "dev_proto_attr_types",   dev_proto_attr_types)
  dserv(b, ctx, "dev_proto_data_types",   dev_proto_data_types)

  aserv(b, ctx, "dev_proto_attr_list",    dev_proto_attr_list)
  aserv(b, ctx, "dev_proto_attr_update",  dev_proto_attr_update)
  aserv(b, ctx, "dev_proto_attr_delete",  dev_proto_attr_delete)

  aserv(b, ctx, "dev_proto_data_list",    dev_proto_data_list)
  aserv(b, ctx, "dev_proto_data_update",  dev_proto_data_update)
  aserv(b, ctx, "dev_proto_data_delete",  dev_proto_data_delete)

  aserv(b, ctx, "dev_proto_ctrl_list",    dev_proto_ctrl_list)
  aserv(b, ctx, "dev_proto_ctrl_update",  dev_proto_ctrl_update)
  aserv(b, ctx, "dev_proto_ctrl_delete",  dev_proto_ctrl_delete)
}


func dev_proto_count(h *Ht) interface{} {
  return h.Crud().PageInfo()
}


func dev_proto_list(h *Ht) interface{} {
  return h.Crud().List(func(opt *options.FindOptions) {
    opt.SetProjection(bson.M{
      "desc":1, "cd":1, "md":1, "changeid":1, "script":1 })
  })
}


func dev_proto_create(h *Ht) interface{} {
  //TODO: 验证脚本
  scriptName := checkstring("脚本", h.Get("script"), 0, 64)

  d := bson.D{
    {"_id",       checkstring("原型ID", h.Get("id"), 2, 20)},
    {"desc",      checkstring("原型说明", h.Get("desc"), 0, 999)},
    {"cd", 	      time.Now()},
    {"md",        ""},
    {"changeid",  1},
    {"script",    scriptName},
    {"attrs",     []bson.M{} },
    {"datas",     []bson.M{} },
    {"ctrls",     []bson.M{} },
  }
  return h.Crud().Create(&d)
}


func dev_proto_update(h *Ht) interface{} {
  id := checkstring("原型ID", h.Get("id"), 2, 20)
  //TODO: 验证脚本
  scriptName := checkstring("脚本", h.Get("script"), 0, 64)

  up := bson.M{
    "$inc" : bson.M{ "changeid" : 1 },
    "$set" : bson.M{
      "desc"   : h.Get("desc"), 
      "md"     : time.Now(),
      "script" : scriptName,
    },
  }

  return h.Crud().Update(id, up)
}


func dev_proto_delete(h *Ht) interface{} {
  id := checkstring("原型ID", h.Get("id"), 2, 20)

  if GetDeviceRefProto(h.Ctx(), id) {
    return HttpRet{3, "原型被设备引用, 不能删除", id}
  }
  return h.Crud().Delete(id)
}


func dev_proto_read(h *Ht) interface{} {
  id := checkstring("原型ID", h.Get("id"), 2, 20)
  return h.Crud().Read(id)
}


func dev_proto_attr_types(h *Ht) interface{} {
  return core.DAT__map
}


func dev_proto_data_types(h *Ht) interface{} {
  return core.DDT__map
}


func dev_proto_attr_list(h *Ht) interface{} {
  id := checkstring("原型ID", h.Get("id"), 2, 20)
  return h.Crud().Read(id, "attrs")
}


func dev_proto_attr_update(h *Ht) interface{} {
  id   := checkstring("原型ID", h.Get("id"), 2, 20)
  name := checkstring("属性名称", h.Get("name"), 2, 20)

  max  := checkint("最大值", h.Get("max"), INT_MIN, INT_MAX)
  min  := checkint("最小值", h.Get("min"), INT_MIN, INT_MAX)
  typ  := checkint("类型", h.Get("type"), 100, 200)
  dict := h.Get("dict")

  if _, has := core.DAT__map[core.DevAttrType(typ)]; !has {
    return HttpRet{1, "无效的类型", typ}
  }

  if typ == int64(core.DAT_dict) {
    if dict == "" {
      return errors.New("字典不能为空")
    }
  } else {
    if min >= max {
      return errors.New("最小值必须小于最大值")
    }
  }

  fields := bson.M{
    "name"    : name,
    "desc"    : checkstring("属性说明", h.Get("desc"), 0, 99),
    "type"    : typ,
    "notnull" : checkbool("非空", h.Get("notnull")),
    "defval"  : checkstring("默认值", h.Get("defval"), 0, 99),
    "dict"    : dict,
    "max"     : max,
    "min"     : min,
  }

  up := bson.M{ 
    "$set" : bson.M{ "md" : time.Now() },
    "$inc" : bson.M{"changeid" : 1},
  }
  return h.Crud().UpdateInnerArray(id, "attrs", "name", fields, up)
}


func dev_proto_attr_delete(h *Ht) interface{} {
  return __proto_remove_arr(h, "属性", "attrs")
}


func __proto_remove_arr(h *Ht, info string, attrName string) interface{} {
  id   := checkstring("原型ID", h.Get("id"), 2, 20)
  name := checkstring(info +"名称", h.Get("name"), 2, 20)

  filter := bson.M{ "_id" : id, attrName +".name" : name }
  up     := bson.M{ 
    "$set"   : bson.M{ "md" : time.Now() },
    "$inc"   : bson.M{ "changeid" : 1 },
    "$pull"  : bson.M{ attrName : bson.M{ "name": name } },
  }
  _, err := h.Table().UpdateOne(h.Ctx(), filter, up)
  if err != nil {
    return err
  }
  return HttpRet{0, info +"已删除", nil}
}


func dev_proto_data_list(h *Ht) interface{} {
  id := checkstring("原型ID", h.Get("id"), 2, 20)
  return h.Crud().Read(id, "datas")
}


func dev_proto_data_update(h *Ht) interface{} {
  id   := checkstring("原型ID", h.Get("id"), 2, 20)
  name := checkstring("名称", h.Get("name"), 2, 20)
  typ  := checkint("类型", h.Get("type"), 1, 100)

  if _, has := core.DDT__map[core.DevDataType(typ)]; !has {
    return HttpRet{1, "无效的类型", typ}
  }

  fields := bson.M{
    "name"    : name,
    "desc"    : checkstring("说明", h.Get("desc"), 0, 99),
    "type"    : typ,
  }

  up := bson.M{ 
    "$set" : bson.M{ "md" : time.Now() },
    "$inc" : bson.M{ "changeid" : 1 },
  }

  return h.Crud().UpdateInnerArray(id, "datas", "name", fields, up)
}


func dev_proto_data_delete(h *Ht) interface{} {
  return __proto_remove_arr(h, "数据槽", "datas")
}


func dev_proto_ctrl_list(h *Ht) interface{} {
  id := checkstring("原型ID", h.Get("id"), 2, 20)
  return h.Crud().Read(id, "ctrls")
}


func dev_proto_ctrl_update(h *Ht) interface{} {
  id   := checkstring("原型ID", h.Get("id"), 2, 20)
  name := checkstring("名称", h.Get("name"), 2, 20)
  typ  := checkint("类型", h.Get("type"), 1, 100)

  if _, has := core.DDT__map[core.DevDataType(typ)]; !has {
    return HttpRet{1, "无效的类型", typ}
  }

  fields := bson.M{
    "name"    : name,
    "desc"    : checkstring("说明", h.Get("desc"), 0, 99),
    "type"    : typ,
  }

  up := bson.M{ 
    "$set" : bson.M{ "md" : time.Now() },
    "$inc" : bson.M{"changeid" : 1},
  }

  return h.Crud().UpdateInnerArray(id, "ctrls", "name", fields, up)
}


func dev_proto_ctrl_delete(h *Ht) interface{} {
  return __proto_remove_arr(h, "控制槽", "ctrls")
}


//
// 返回原型数据
//
func GetDevProto(ctx context.Context, id string, ret interface{}) error {
  filter := bson.M{"_id": id}
  return mg.Collection(core.TableDevProto).FindOne(ctx, filter).Decode(ret)
}
