package service

import (
	"ic1101/brick"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//
// 用于增删改查的便捷操作
//
type Crud struct {
  h         brick.Http
  collname  string
  info      string
}

//
// 立即创建数据并返回
//
func (c* Crud) Create(data interface{}) error {
  table := mg.Collection(c.collname)
  _, err := table.InsertOne(c.h.Ctx(), data) 
  if err != nil {
    c.h.Json(HttpRet{1, c.info +"创建失败(id重复)", err.Error()})
    return nil
  }
  c.h.Json(HttpRet{0, c.info +"已创建", nil})
  return nil
}

//
// 立即返回分页数据
//
func (c* Crud) PageInfo() error {
  table := mg.Collection(c.collname)
  count, _ := table.CountDocuments(c.h.Ctx(), bson.D{})
  pageret := struct {
    Count int64
    PageSize int64
  }{count, PageSize}
  c.h.Json(HttpRet{0, "pageinfo", pageret})
  return nil
}

//
// 立即返回查询数据, 
// text 为全文检索 uri 参数
// page 为分页号码(0开始)
// 必须创建全文检索索引
//
func (c* Crud) List(set_options func(*options.FindOptions)) error {
  fo := options.Find()
  fo.SetLimit(PageSize)
  fo.SetSkip(checkpage(c.h))

  if set_options != nil {
    set_options(fo)
  }

  filter := bson.M{}
  t := c.h.Get("text")
  if t != "" {
    filter["$text"] = bson.D{{"$search", t}}
  }
  
  table := mg.Collection(c.collname)
  cursor, err := table.Find(c.h.Ctx(), filter, fo)

  if err != nil {
    c.h.Json(HttpRet{1, "查询"+ c.info +"错误", err.Error()})
    return nil
  }

  var results []bson.M
  cursor.All(c.h.Ctx(), &results)
  c.h.Json(HttpRet{0, "查询"+ c.info, &results})
  return nil
}

//
// 立即删除一行数据, 必须有 id 属性
//
func (c *Crud) Delete(id string) error {
  table := mg.Collection(c.collname)
  _, err := table.DeleteOne(c.h.Ctx(), bson.D{{"_id", id}})
  if err != nil {
    c.h.Json(HttpRet{1, c.info +"删除错误", err.Error()})
  } else {
    c.h.Json(HttpRet{0, c.info +"已删除", id})
  }
  return nil
}

//
// 立即更新一行数据, data 是完整的更新命令
// md 属性总是保存更新日期
//
func (c *Crud) Update(id string, data interface{}) error {
  table  := mg.Collection(c.collname)
  filter := bson.D{{"_id", id}}

  if _, err := table.UpdateOne(c.h.Ctx(), filter, data); err != nil {
    c.h.Json(HttpRet{1, c.info +"更新错误", err.Error()})
    return nil
  }
  c.h.Json(HttpRet{0, c.info +"已更新", id})
  return nil
}


//
// 立即返回一行数据
//
func (c *Crud) Read(id string, includeNames ...string) error {
  table  := mg.Collection(c.collname)
  filter := bson.D{{"_id", id}}
  ret    := bson.M{}
  opt    := options.FindOne()
  
  if len(includeNames) > 0 {
    proj := bson.M{}
    for _, name := range includeNames {
      proj[name] = 1
    }
    opt.SetProjection(proj)
  }

  if err := table.FindOne(c.h.Ctx(), filter, opt).Decode(&ret); err != nil {
    c.h.Json(HttpRet{1, c.info +"查询错误", err.Error()})
    return nil
  }
  c.h.Json(HttpRet{0, "返回"+ c.info, ret})
  return nil
}