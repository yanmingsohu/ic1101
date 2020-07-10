package service

import (
	"context"
	"errors"
	"ic1101/brick"
	"ic1101/src/core"
	"io/ioutil"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

var __i uint8 = 0
var __x uint8 = 0
var __e error
var __c = make(chan uint8)
var license_id = "license"
var license_filter = bson.M{ "_id": license_id }



func installLicenseService(b *brick.Brick) {
  ctx := &ServiceGroupContext{core.TableSystem, "软件授权"}
  __start__license(ctx)

  lserv(b, ctx, "license_get_info",  license_get_info)
  lserv(b, ctx, "license_get_req",   license_get_req)
  lserv(b, ctx, "license_update",    license_update)
  dserv(b, ctx, "license_get_state", license_get_state)
}


func __start__license(ctx *ServiceGroupContext) {
  coll := mg.Collection(ctx.collectionName)

  look_mem := func() {
    li := core.Li{}
    err := coll.FindOne(context.Background(), license_filter).Decode(&li)
    if err != nil {
      __e = err
      return
    }
    
    li.ComputeZ()
    err = li.Verification()
    if err == nil {
      err = li.CheckTime()
    }
    if err == nil {
      __i = uint8(rand.Int() % 0xEF)
      __x = __i + 1
    } else {
      __i = 1
      __x = 0
      __e = err
    }
  }

  go (func() {
    for _ = range __c {
      // log.Println("验证授权")
      look_mem()
    }
  })()

  go (func() {
    tk := time.NewTicker(time.Hour * 3)
    for range tk.C {
      __c <- 0
    }
  })()
  
  // 系统启动时检查一次
  look_mem()
}


func license_get_state(h *Ht) interface{} {
  if __x <= __i {
    return HttpRet{1, "", nil}
  }
  return HttpRet{0, "", nil}
}


func license_get_info(h *Ht) interface{} {
  li := core.Li{}
  if err := h.Table().FindOne(h.Ctx(), license_filter).Decode(&li); err != nil {
    li.AppName   = core.GAppName
    li.BeginTime = _begin_time()
    li.EndTime   = _end_time()
  }
  return HttpRet{0, "", li}
}


func license_get_req(h *Ht) interface{} {
  li := _paramter_license(h)
  li.ComputeZ()
  yaml, err := li.String()
  if err != nil {
    return err
  }
  return HttpRet{0, "", yaml}
}


func license_update(h *Ht) interface{} {
  body, err := ioutil.ReadAll(h.R.Body)
  if err != nil {
    return err
  }
  if len(body) < 10 {
    return errors.New("无效的递交参数")
  }
  defer (func() {
    __c <- 0
  })()
  li := core.Li{}
  li.Init(string(body))
  return h.Crud().Upsert(license_id, bson.M{"$set": li})
}


func _paramter_license(h *Ht) core.Li {
  li := core.Li{
    AppName   : checkstring("应用名称", h.Get("appName"), 5, 99),
    Company   : checkstring("授权单位", h.Get("company"), 3, 99),
    Dns       : h.Get("dns"),
    Email     : h.Get("email"),
    BeginTime : h.GetUint64("beginTime", _begin_time()),
    EndTime   : h.GetUint64("endTime", _end_time()),
  }
  return li
}


// 当前日期的前一天
func _begin_time() uint64 {
  return uint64(time.Now().Unix() - 24*60*60) * 1000
}


// 无限使用
func _end_time() uint64 {
  return uint64(time.Now().Unix() + 100*356*24*60*60) * 1000
}