/**
 *  Copyright 2023 Jing Yanming
 * 
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */
package service

import (
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//
// 用于增删改查的便捷操作
//
type Crud struct {
  h         *Ht
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
  res, err := table.DeleteOne(c.h.Ctx(), bson.D{{"_id", id}})
  
  if err != nil {
    c.h.Json(HttpRet{1, c.info +"删除错误", err.Error()})
  }
  if res.DeletedCount < 1 {
    c.h.Json(HttpRet{1, c.info +"数据不存在", nil})
    return nil
  }

  c.h.Json(HttpRet{0, c.info +"已删除", id})
  return nil
}

//
// 立即更新一行数据, data 是完整的更新命令
// md 属性总是保存更新日期
//
func (c *Crud) Update(id string, data interface{}, opts ...*options.UpdateOptions) error {
  table  := mg.Collection(c.collname)
  filter := bson.D{{"_id", id}}

  res, err := table.UpdateOne(c.h.Ctx(), filter, data, opts...)
  if err != nil {
    c.h.Json(HttpRet{1, c.info +"更新错误", err.Error()})
    return nil
  }

  if res.UpsertedCount > 0 {
    c.h.Json(HttpRet{0, c.info +" 已创建", id})
  } else {
    if res.MatchedCount < 1 {
      c.h.Json(HttpRet{1, c.info +"数据不存在", nil})
      return nil
    }
    c.h.Json(HttpRet{0, c.info +" 已更新", id})
  }
  return nil
}


//
// 更新或插入数据
//
func (c *Crud) Upsert(id string, data interface{}) error {
  opt := options.Update()
  opt.SetUpsert(true)
  return c.Update(id, data, opt)
}


//
// 立即返回一行数据到 http 接口
//
func (c *Crud) Read(id string, includeNames ...string) error {
  ret, err := c.DRead(id, includeNames...)
  if err != nil {
    c.h.Json(HttpRet{1, c.info +"查询错误", err.Error()})
    return nil
  }
  c.h.Json(HttpRet{0, "返回"+ c.info, ret})
  return nil
}


//
// 返回查询的数据
//
func (c *Crud) DRead(id string, includeNames ...string) (bson.M, error) {
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
    return nil, err
  }
  return ret, nil
}


//
// 更新文档中类型为对象数组列表中的元素, listName 为对象列表的名字. 
// attrName 是列表中的一个元素的属性名, 当找到匹配时使用 fields 更新元素;
// 如果数组中没有匹配元素则在数组中插入新的 fields 元素.
// id     -- 主键
// fields -- 数组元素
// up     -- 更新参数
//
func (c *Crud) UpdateInnerArray(id string, listName string, 
      attrName string, fields bson.M, up bson.M) interface{} {

  attrVal := fields[attrName].(string)
  set_array := listName +".$"
  filter := bson.M{ 
    "_id" : id, 
    listName +"."+ attrName : attrVal,
  }
  up["$set"].(bson.M)[set_array] = fields

  log.Print("UPDATE", up)
  ur, err := c.h.Table().UpdateOne(c.h.Ctx(), filter, up)
  if err != nil {
    return err
  }

  if ur.MatchedCount == 0 {
    delete(up["$set"].(bson.M), set_array);
    up["$push"] = bson.M{ listName : fields }
    log.Print("INSERT", up)

    filter := bson.M{ "_id" : id }
    _, err = c.h.Table().UpdateOne(c.h.Ctx(), filter, up)

    if err != nil {
      return err
    }
    return HttpRet{0, attrVal +"已创建", nil}
  }

  return HttpRet{0, attrVal +"已更新", nil}
}