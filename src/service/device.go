package service

import (
	"context"
	"ic1101/brick"
	"ic1101/src/core"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)


func installDeviceService(b *brick.Brick) {  
  mg.CreateIndex(core.TableDevice, &bson.D{{"_id", "text"}, {"desc", "text"}})
  mg.CreateIndex(core.TableDevice, &bson.D{{"tid", 1}})
  ctx := &ServiceGroupContext{core.TableDevice, "设备"}

  aserv(b, ctx, "dev_count",    dev_count)
  aserv(b, ctx, "dev_list",     dev_list)
  aserv(b, ctx, "dev_upsert",   dev_upsert)
  aserv(b, ctx, "dev_delete",   dev_delete)
  aserv(b, ctx, "dev_read",     dev_read)
}


func dev_count(h *Ht) interface{} {
  return h.Crud().PageInfo()
}


func dev_list(h *Ht) interface{} {
  return h.Crud().List(func(opt *options.FindOptions) {
    opt.SetProjection(bson.M{ "attrs":0 })
  }) 
}


func dev_read(h *Ht) interface{} {
  id := checkstring("设备ID", h.Get("id"), 2, 20)
  return h.Crud().Read(id)
}


func dev_upsert(h *Ht) interface{} {
  id   := checkstring("设备ID", h.Get("id"), 2, 20)
  desc := checkstring("设备说明", h.Get("desc"), 0, 999)
  var tid string
  
  exists := bson.M{}
  if nil == GetDevice(h.Ctx(), id, exists) {
    tid = exists["tid"].(string)
  } else {
    tid = checkstring("设备原型ID", h.Get("tid"), 2, 20)
  }

  proto := core.DevProto{}
  if err := GetDevProto(h.Ctx(), tid, &proto); err != nil {
    return HttpRet{1, "不存在的设备原型 "+ tid, err.Error()}
  }

  dev_attrs := bson.M{}

  for _, item := range proto.Attrs {
    attrName := item.Name
    attrVal := h.Get("a."+ attrName)

    if attrVal == "" {
      if item.Notnull {
        return HttpRet{2, "参数 '"+ attrName +"' 不能为空", attrName}
      } else {
        dev_attrs[attrName] = nil
        continue;
      }
    }

    switch (item.Type) {
    case core.DAT_date:
      time, err := time.Parse(time.RFC1123, attrVal);
      if err != nil {
        return err
      }
      dev_attrs[attrName] = time
      break;

    case core.DAT_dict:
      if !hasKeyInDict(h.Ctx(), item.Dict, attrVal) {
        return HttpRet{3, "字典值不在字典中 "+ attrVal, attrVal}
      }
      dev_attrs[attrName] = attrVal
      break;

    case core.DAT_number:
      dev_attrs[attrName] = 
          checkint(attrName, attrVal, item.Min, item.Max)
      break;

    case core.DAT_string:
      dev_attrs[attrName] = 
          checkstring(attrName, attrVal, int(item.Min), int(item.Max))
      break;

    default:
      return HttpRet{3, "无效的参数类型", item.Type}
    }
  }
  
  d := bson.M{
    "_id"       : id,
    "desc"      : desc,
    "tid"       : proto.Id,
    "changeid"  : proto.ChangeId,
    "md"        : "",
    "attrs"     : dev_attrs,
  }
  if _, has := exists["_id"]; has {
    d["md"] = time.Now()
  } else {
    d["cd"] = time.Now()
    d["dd"] = ""
    d["dc"] = 0
  }
  return h.Crud().Upsert(id, bson.M{"$set" : d})
}


func dev_delete(h *Ht) interface{} {
  //TODO: 不能删除有数据的设备, 不能删除挂接在总线上的设备
  id := checkstring("设备ID", h.Get("id"), 2, 20)
  return h.Crud().Delete(id)
}


//
// 返回设备数据, 失败返回 error, 成功数据填入 ret 中.
//
func GetDevice(ctx context.Context, id string, ret interface{}) error {
  filter := bson.M{ "_id": id }
  return mg.Collection(core.TableDevice).FindOne(ctx, filter).Decode(ret)
}


//
// 查询引用原型的设备, 至少有一个对原型的引用返回 true.
//
func GetDeviceRefProto(ctx context.Context, protoid string) bool {
  cur, err := mg.Collection(core.TableDevice).Find(ctx, 
      bson.M{"tid": protoid}, options.Find().SetLimit(1))
  if err != nil {
    panic(err)
  }
  defer cur.Close(ctx)
  return cur.Next(ctx)
}