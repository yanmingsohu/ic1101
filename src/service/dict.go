package service

import (
	"errors"
	"time"

	"ic1101/brick"
	"ic1101/src/core"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)


func installDictService(b *brick.Brick) {
  aserv(b, "dict_create",  dict_create)
  aserv(b, "dict_read",    dict_read)
  aserv(b, "dict_sets",    dict_sets)
  aserv(b, "dict_list",    dict_list)
  aserv(b, "dict_count",   dict_count)
  aserv(b, "dict_delete",  dict_delete)
}


func dict_create(h brick.Http) error {
  d := bson.D{
    {"_id",  checkstring("字典ID", h.Get("id"), 2, 20)},
    {"desc", checkstring("字典说明", h.Get("desc"), 0, 999)},
    {"cd", 	 time.Now()},
    {"md",   ""},
  }

  table := mg.Collection("dict")
  _, err := table.InsertOne(h.Ctx(), d) 
  if err != nil {
    h.Json(HttpRet{1, "字典创建失败(id重复)", err})
    return nil
  }
  h.Json(HttpRet{0, "字典已创建", nil})
  return nil
}


func dict_count(h brick.Http) error {
  table := mg.Collection("dict")
  count, _ := table.CountDocuments(h.Ctx(), bson.D{})
  pageret := struct {
    Count int64
    PageSize int64
  }{count, PageSize}
  h.Json(HttpRet{0, "pageinfo", pageret})
  return nil
}


func dict_list(h brick.Http) error {
  fo := options.Find()
  fo.SetLimit(PageSize)
  fo.SetSkip(checkpage(h))
  fo.SetProjection(bson.M{"desc":1, "cd":1, "md":1})

  id := h.Get("id")
  desc := h.Get("desc")
  filter := bson.M{}
  if id != "" {
    filter["_id"] = id
  }
  if desc != "" {
    filter["desc"] = desc
  }
  
  table := mg.Collection("dict")
  cursor, err := table.Find(h.Ctx(), filter, fo)

  if err != nil {
    h.Json(HttpRet{1, "查询错误", err})
    return nil
  }

  var results []bson.M
  cursor.All(h.Ctx(), &results)
  h.Json(HttpRet{0, "list", &results})
  return nil
}


func dict_read(h brick.Http) error {
  id     := checkstring("字典ID", h.Get("id"), 2, 20)
  table  := mg.Collection("dict")
  filter := bson.D{{"_id", id}}
  dict   := core.Dict{}
  if err := table.FindOne(h.Ctx(), filter).Decode(&dict); err != nil {
    return errors.New("字典不存在")
  }
  h.Json(HttpRet{0, "返回字典", &dict})
  return nil
}


func dict_sets(h brick.Http) error {
  id   := checkstring("字典ID", h.Get("id"), 2, 20)
  desc := h.Get("desc")
  kv   := map[string]string{}
  keys := h.Gets("k")
  vs   := h.Gets("v")

  for i, k := range keys {
    if k == "" {
      return errors.New("不允许属性名为空")
    }
    if vs[i] == "" {
      return errors.New("属性 "+ k +" 的值为空")
    }
    kv[k] = vs[i]
  }

  table  := mg.Collection("dict")
  filter := bson.D{{"_id", id}}
  up     := bson.D{{"$set", bson.D{{"desc", desc}, {"content", kv}} }}

  if _, err := table.UpdateOne(h.Ctx(), filter, up); err != nil {
    return errors.New("字典不存在")
  }
  h.Json(HttpRet{0, "字典已更新", id})
  return nil
}


func dict_delete(h brick.Http) error {
  id := checkstring("字典ID", h.Get("id"), 2, 20)
  table := mg.Collection("dict")
  _, err := table.DeleteOne(h.Ctx(), bson.D{{"_id", id}})
  if err != nil {
    h.Json(HttpRet{1, "字典删除错误", err})
  } else {
    h.Json(HttpRet{0, "字典已删除", id})
  }
  return nil
}