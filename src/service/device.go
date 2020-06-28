package service

import (
	"ic1101/brick"
	"ic1101/src/core"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)


func installDeviceService(b *brick.Brick) {  
  mg.CreateIndex(core.TableDevice, &bson.D{{"_id", "text"}, {"desc", "text"}})
  ctx := &ServiceGroupContext{core.TableDevice, "设备"}

  aserv(b, ctx, "dev_count",   dev_count)
  aserv(b, ctx, "dev_list",    dev_list)
  aserv(b, ctx, "dev_upsert",  dev_upsert)
  aserv(b, ctx, "dev_delete",  dev_delete)
}


func dev_count(h *Ht) interface{} {
  return h.Crud().PageInfo()
}


func dev_list(h *Ht) interface{} {
  return h.Crud().List(func(opt *options.FindOptions) {
    opt.SetProjection(bson.M{
      "desc":1, "tid":1, "changeid":1, "md":1, "cd":1, "dd":1, "dc":1 })
  }) 
}


func dev_upsert(h *Ht) interface{} {
  id   := checkstring("设备ID", h.Get("id"), 2, 20)
  desc := checkstring("设备说明", h.Get("desc"), 0, 999)
  tid  := checkstring("设备原型ID", h.Get("id"), 2, 20)

  proto := bson.M{}
  if err := getDevProto(h.Ctx(), tid, proto); err != nil {
    return HttpRet{1, "不存在的设备原型 "+ tid, err}
  }

  dev_attrs := bson.M{}
  proto_attrs := proto["attrs"].([]bson.M)

  for _, item := range proto_attrs {
    attrName := item["name"].(string)
    attrVal := h.Get("a."+ attrName)

    if attrVal == "" {
      if item["notnull"].(bool) {
        return HttpRet{2, "参数"+ attrName +"不能为空", attrName}
      } else {
        continue;
      }
    }

    switch (item["type"].(core.DevAttrType)) {
    case core.DAT_date:
      time, err := time.Parse(core.TimeFormatString, attrVal);
      if err != nil {
        return err
      }
      dev_attrs[attrName] = time
      break;

    case core.DAT_dict:
      if !hasKeyInDict(h.Ctx(), item["dict"].(string), attrVal) {
        return HttpRet{3, "字典值不在字典中 "+ attrVal, attrVal}
      }
      dev_attrs[attrName] = attrVal
      break;

    case core.DAT_number:
      min := item["min"].(int64)
      max := item["max"].(int64)
      dev_attrs[attrName] = checkint(attrName, attrVal, min, max)
      break;

    case core.DAT_string:
      min := item["min"].(int)
      max := item["max"].(int)
      dev_attrs[attrName] = checkstring(attrName, attrVal, min, max)
      break;

    default:
      return HttpRet{3, "无效的参数类型", item["type"]}
    }
  }
  
  d := bson.M{
    "_id"       : id,
    "desc"      : desc,
    "tid"       : proto["_id"],
    "changeid"  : proto["changeid"],
    "md"        : "",
    "cd"        : time.Now(),
    "dd"        : "",
    "dc"        : 0,
    "attrs"     : dev_attrs,
    "data_years": bson.M{},
  }
  return h.Crud().Upsert(id, d)
}


func dev_delete(h *Ht) interface{} {
  return nil
}