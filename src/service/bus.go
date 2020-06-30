package service

import (
	"errors"
	"ic1101/brick"
	"ic1101/src/bus"
	"ic1101/src/core"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)


func installBusService(b *brick.Brick) {
  mg.CreateIndex(core.TableDevice, &bson.D{{"_id", "text"}, {"desc", "text"}})
  ctx := &ServiceGroupContext{core.TableBus, "总线"}

  aserv(b, ctx, "bus_count",  bus_count)
  aserv(b, ctx, "bus_list",   bus_list)
  aserv(b, ctx, "bus_delete", bus_delete)
  aserv(b, ctx, "bus_create", bus_create)
  aserv(b, ctx, "bus_update", bus_update)
  
  aserv(b, ctx, "bus_types",  bus_types)
}


func bus_count(h *Ht) interface{} {
  return h.Crud().PageInfo()
}


func bus_list(h *Ht) interface{} {
  return h.Crud().List(func (o *options.FindOptions) {
    o.SetProjection(bson.M{
      "data_slot":0, "ctrl_slot":0,
    })
  })
}


func bus_delete(h *Ht) interface{} {
  id := checkstring("总线ID", h.Get("id"), 2, 20)
  if bus.BusStateStop != bus.GetBusState(id) {
    return HttpRet{3, "总线正在运行, 不能删除", id}
  }
  return h.Crud().Delete(id)
}


func bus_create(h *Ht) interface{} {
  id  := checkstring("总线ID", h.Get("id"), 2, 20)
  typ := checkstring("总线类型", h.Get("type"), 2, 20)
  tm  := checkstring("定时器", h.Get("timer"), 2, 20)
  
  if !bus.HasTypeName(typ) {
    return errors.New("无效的总线类型")
  }
  
  if !hasTimer(tm) {
    return errors.New("定时器不存在 "+ tm)
  }

  d := bson.M{
    "_id"       : id,
    "desc"      : checkstring("总线说明", h.Get("desc"), 0, 99),
    "timer"     : tm,
    "cd"        : time.Now(),
    "md"        : "",
    "type"      : typ,
    "status"    : bus.BusStateStop,
    "data_slot" : bson.M{},
    "ctrl_slot" : bson.M{},
  }
  return h.Crud().Create(d)
}


func bus_types(h *Ht) interface{} {
  return HttpRet{0, "类型列表", bus.GetTypes()}
}


func bus_update(h *Ht) interface{} {
  id  := checkstring("总线ID", h.Get("id"), 2, 20)
  tm  := checkstring("定时器", h.Get("timer"), 2, 20)

  if !hasTimer(tm) {
    return errors.New("定时器不存在 "+ tm)
  }

  d := bson.M{
    "desc"      : checkstring("总线说明", h.Get("id"), 2, 20),
    "timer"     : tm,
    "md"        : time.Now(),
    "data_slot" : bson.M{},
    "ctrl_slot" : bson.M{},
  }
  return h.Crud().Update(id, d)
}