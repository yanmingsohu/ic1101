package service

import (
	"context"
	"errors"
	"time"

	"ic1101/brick"
	"ic1101/src/core"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)


func installDictService(b *brick.Brick) {
  ctx := &ServiceGroupContext{core.TableDict, "字典"}
  aserv(b, ctx, "dict_create",     dict_create)
  aserv(b, ctx, "dict_read",       dict_read)
  aserv(b, ctx, "dict_update",     dict_update)
  aserv(b, ctx, "dict_list",       dict_list)
  aserv(b, ctx, "dict_count",      dict_count)
  aserv(b, ctx, "dict_delete",     dict_delete)
  aserv(b, ctx, "dict_insert_key", dict_insert_key)
  aserv(b, ctx, "dict_delete_key", dict_delete_key)

  mg.CreateIndex(core.TableDict, &bson.D{{"_id", "text"}, {"desc", "text"}})
}


func dict_create(h *Ht) interface{} {
  d := bson.D{
    {"_id",  checkstring("字典ID", h.Get("id"), 2, 20)},
    {"desc", checkstring("字典说明", h.Get("desc"), 0, 999)},
    {"cd", 	 time.Now()},
    {"md",   ""},
  }

  return h.Crud().Create(&d)
}


func dict_count(h *Ht) interface{} {
  return h.Crud().PageInfo()
}


func dict_list(h *Ht) interface{} {
  return h.Crud().List(func(opt *options.FindOptions) {
    opt.SetProjection(bson.M{"desc":1, "cd":1, "md":1})
  })
}


func dict_read(h *Ht) interface{} {
  id     := checkstring("字典ID", h.Get("id"), 2, 20)
  filter := bson.D{{"_id", id}}
  dict   := bson.M{}
  opt    := options.Find()
  opt.SetProjection(bson.M{"content":1})

  if err := h.Table().FindOne(h.Ctx(), filter).Decode(&dict); err != nil {
    h.Json(HttpRet{1, "错误", err.Error()})
    return nil
  }
  h.Json(HttpRet{0, "返回字典", dict["content"]})
  return nil
}


func dict_insert_key(h *Ht) interface{} {
  id   := checkstring("字典ID", h.Get("id"), 2, 20)
  keys := h.Gets("k")
  vs   := h.Gets("v")

  filter := bson.D{{"_id", id}}

  for i, k := range keys {
    if k == "" {
      return errors.New("不允许属性名为空")
    }
    if vs[i] == "" {
      return errors.New("属性 "+ k +" 的值为空")
    }

    up := bson.D{{"$set", bson.D{{"content."+ k, vs[i]}} }}
    if _, err := h.Table().UpdateOne(h.Ctx(), filter, up); err != nil {
      return err
    }
  }

  up := bson.D{{"$set", bson.D{{"md", time.Now()}} }}
  if _, err := h.Table().UpdateOne(h.Ctx(), filter, up); err != nil {
    return err
  }
  h.Json(HttpRet{0, "字典已更新", id})
  return nil
}


func dict_delete_key(h *Ht) interface{} {
  id  := checkstring("字典ID", h.Get("id"), 2, 20)
  key := checkstring("属性名", h.Get("k"), 1, 99)

  filter := bson.D{{"_id", id}}
  up     := bson.D{{"$unset", bson.D{{"content."+ key, 1}} }}

  if _, err := h.Table().UpdateOne(h.Ctx(), filter, up); err != nil {
    return err
  }
  return HttpRet{0, "字典已更新", id}
}


func dict_update(h *Ht) interface{} {
  id := checkstring("字典ID", h.Get("id"), 2, 20)
  up := bson.D{{"$set", 
        bson.D{{"desc", h.Get("desc")}, {"md", time.Now()}} }}

  return h.Crud().Update(id, up)
}


func dict_delete(h *Ht) interface{} {
  id := checkstring("字典ID", h.Get("id"), 2, 20)
  return h.Crud().Delete(id)
}


func hasKeyInDict(ctx context.Context, dict string, key string) (bool) {
  res := mg.Collection(core.TableDict).FindOne(
      ctx, bson.M{
        "_id": dict, 
        "content."+ key : bson.M{ "$exists": true },
      }, options.FindOne().SetProjection(bson.M{"_id":1}))

  if res.Err() != nil {
    return false
  }
  return true
}