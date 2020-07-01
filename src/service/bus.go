package service

import (
	"context"
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
  
  aserv(b, ctx, "bus_types",        bus_types)
  aserv(b, ctx, "bus_slot_list",    bus_slot_list)
  aserv(b, ctx, "bus_slot_update",  bus_slot_update)
  aserv(b, ctx, "bus_slot_delete",  bus_slot_delete)
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
  if bus.GetBusState(id) != bus.BusStateStop {
    return errors.New("总线运行中, 禁止修改 "+ id)
  }
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

  if bus.GetBusState(id) != bus.BusStateStop {
    return errors.New("总线运行中, 禁止修改 "+ id)
  }
  if !hasTimer(tm) {
    return errors.New("定时器不存在 "+ tm)
  }

  d := bson.M{
    "desc"      : checkstring("总线说明", h.Get("desc"), 0, 99),
    "timer"     : tm,
    "md"        : time.Now(),
  }
  return h.Crud().Update(id, bson.M{"$set": d})
}


func bus_slot_list(h *Ht) interface{} {
  id      := checkstring("总线ID", h.Get("id"), 2, 20)
  isData  := h.GetBool("isdata")
  bus     := bson.M{}
  if err := GetBus(id, h.Ctx(), &bus); err != nil {
    return err
  }
  if isData {
    return HttpRet{0, "总线数据槽列表", bus["data_slot"]}
  } else {
    return HttpRet{0, "总线控制槽列表", bus["ctrl_slot"]}
  }
}


func bus_slot_update(h *Ht) interface{} {
  id        := checkstring("总线ID", h.Get("id"), 2, 20)
  slot_id   := checkstring("总线端口", h.Get("slot_id"), 2, 20)
  dev_id    := checkstring("设备ID", h.Get("dev_id"), 2, 20)
  data_name := checkstring("设备数据名", h.Get("data_name"), 2, 20)

  if bus.GetBusState(id) != bus.BusStateStop {
    return errors.New("总线运行中, 禁止修改 "+ id)
  }
  findbus := core.Bus{}
  if err := GetBus(id, h.Ctx(), &findbus); err != nil {
    return err
  }
  sp , err := bus.GetSlotParser(findbus.Type)
  if err != nil {
    return err
  }
  slot, err := sp.ParseSlot(slot_id)
  if err != nil {
    return err
  }

  dev := core.Device{}
  if err := GetDevice(h.Ctx(), dev_id, &dev); err != nil {
    return err
  }
  proto := core.DevProto{}
  if err := GetDevProto(h.Ctx(), dev.ProtoId, &proto); err != nil {
    return err
  }
  var ppdarr []core.DevProtoData
  if slot.Type() == bus.SlotData {
    ppdarr = proto.Datas
  } else {
    ppdarr = proto.Ctrls
  }
  pd, err := core.FindProtoDataByName(ppdarr, data_name)
  if err != nil {
    return err
  }

  slot_conf := bson.M{
    "slot_id"   : slot_id,
    "slot_desc" : slot.Desc(),
    "dev_id"    : dev_id,
    "data_name" : data_name,
    "data_type" : pd.Type,
    "data_desc" : pd.Desc,
  }
  set := bson.M{
    "md" : time.Now(),
  }

  if slot.Type() == bus.SlotData {
    set["data_slot."+ slot_id] = slot_conf
  } else {
    timer_id := checkstring("定时器ID", h.Get("ctrl_timer"), 2, 20)
    if !hasTimer(timer_id) {
      return HttpRet{6, "定时器无效 "+ timer_id, timer_id}
    }

    value, err := pd.Type.Parse( h.Get("value") )
    if err != nil {
      return HttpRet{5, "无法解析参数", err}
    }

    slot_conf["timer"] = timer_id
    slot_conf["value"] = value
    set["ctrl_slot."+ slot_id] = slot_conf
  }
  return h.Crud().Update(id, bson.M{ "$set": set })
}


func bus_slot_delete(h *Ht) interface{} {
  id        := checkstring("总线ID", h.Get("id"), 2, 20)
  slot_id   := checkstring("端口", h.Get("slot_id"), 2, 20)
  isdata    := h.GetBool("isdata")

  if bus.GetBusState(id) != bus.BusStateStop {
    return errors.New("总线运行中, 禁止修改 "+ id)
  }
  
  var key string
  if isdata {
    key = "data_slot."+ slot_id
  } else {
    key = "ctrl_slot."+ slot_id
  }
  
  up := bson.M{
    "$set"   : bson.M{ "md" : time.Now() },
    "$unset" : bson.M{ key : true },
  }
  filter := bson.M{ 
    "_id" : id,
    key   : bson.M{ "$exists": true },
  }
  r, err := h.Table().UpdateOne(h.Ctx(), filter, up)
  if err != nil {
    return err
  }
  
  if r.MatchedCount < 1 {
    return errors.New("数据不存在")
  }
  if isdata {
    return HttpRet{0, "数据槽已删除", slot_id}
  } else {
    return HttpRet{0, "控制槽已删除", slot_id}
  }
}


func GetBus(id string, ctx context.Context, ret interface{}) error {
  return mg.GetOne(core.TableBus, ctx, id, ret)
}