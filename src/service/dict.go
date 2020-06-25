package service

import (
	"errors"
	"time"

	"ic1101/brick"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)


func installDictService(b *brick.Brick) {
  aserv(b, "dict_create",     dict_create)
  aserv(b, "dict_read",       dict_read)
  aserv(b, "dict_update",     dict_update)
  aserv(b, "dict_list",       dict_list)
  aserv(b, "dict_count",      dict_count)
  aserv(b, "dict_delete",     dict_delete)
  aserv(b, "dict_insert_key", dict_insert_key)

  mg.CreateIndex("dict", &bson.D{{"_id", "text"}, {"desc", "text"}})
}


func dict_create(h brick.Http) error {
  d := bson.D{
    {"_id",  checkstring("字典ID", h.Get("id"), 2, 20)},
    {"desc", checkstring("字典说明", h.Get("desc"), 0, 999)},
    {"cd", 	 time.Now()},
    {"md",   ""},
  }

  c := Crud{h, "dict", "字典"}
  return c.Create(&d)
}


func dict_count(h brick.Http) error {
  c := Crud{h, "dict", "字典"}
  return c.PageInfo()
}


func dict_list(h brick.Http) error {
  c := Crud{h, "dict", "字典"}
  return c.List(func(opt *options.FindOptions) {
    opt.SetProjection(bson.M{"desc":1, "cd":1, "md":1})
  })
}


func dict_read(h brick.Http) error {
  id     := checkstring("字典ID", h.Get("id"), 2, 20)
  table  := mg.Collection("dict")
  filter := bson.D{{"_id", id}}
  dict   := bson.M{}
  opt    := options.Find()
  opt.SetProjection(bson.M{"content":1})

  if err := table.FindOne(h.Ctx(), filter).Decode(&dict); err != nil {
    h.Json(HttpRet{1, "错误", err.Error()})
    return nil
  }
  h.Json(HttpRet{0, "返回字典", dict["content"]})
  return nil
}


func dict_insert_key(h brick.Http) error {
  id   := checkstring("字典ID", h.Get("id"), 2, 20)
  keys := h.Gets("k")
  vs   := h.Gets("v")

  table  := mg.Collection("dict")
  filter := bson.D{{"_id", id}}

  for i, k := range keys {
    if k == "" {
      return errors.New("不允许属性名为空")
    }
    if vs[i] == "" {
      return errors.New("属性 "+ k +" 的值为空")
    }

    up := bson.D{{"$set", bson.D{{"content."+ k, vs[i]}} }}
    if _, err := table.UpdateOne(h.Ctx(), filter, up); err != nil {
      return err
    }
  }

  up := bson.D{{"$set", bson.D{{"md", time.Now()}} }}
  if _, err := table.UpdateOne(h.Ctx(), filter, up); err != nil {
    return err
  }
  h.Json(HttpRet{0, "字典已更新", id})
  return nil
}


func dict_update(h brick.Http) error {
  id := checkstring("字典ID", h.Get("id"), 2, 20)
  up := bson.D{{"$set", 
        bson.D{{"desc", h.Get("desc")}, {"md", time.Now()}} }}

  c := Crud{h, "dict", "字典"}
  return c.Update(id, up)
}


func dict_delete(h brick.Http) error {
  id := checkstring("字典ID", h.Get("id"), 2, 20)
  c := Crud{h, "dict", "字典"}
  return c.Delete(id)
}