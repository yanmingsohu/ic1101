package service

import (
	"context"
	"errors"
	"ic1101/brick"
	"ic1101/src/bus"
	"ic1101/src/core"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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
  
  aserv(b, ctx, "bus_slot_list",    bus_slot_list)
  aserv(b, ctx, "bus_slot_update",  bus_slot_update)
  aserv(b, ctx, "bus_slot_delete",  bus_slot_delete)

  dserv(b, ctx, "bus_types", bus_types)

  aserv(b, ctx, "bus_stop",       bus_stop)
  aserv(b, ctx, "bus_start",      bus_start)
  aserv(b, ctx, "bus_last_data",  bus_last_data)
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


func bus_last_data(h *Ht) interface{} {
  id := checkstring("总线ID", h.Get("id"), 2, 20)
  c  := Crud{h, core.TableBusData, "总线实时数据"}
  return c.Read(id)
}


func bus_stop(h *Ht) interface{} {
  id := checkstring("总线ID", h.Get("id"), 2, 20)
  if err := bus.StopBus(id); err != nil {
    return err
  }
  return HttpRet{0, "总线已终止", nil}
}


func bus_start(h *Ht) interface{} {
  id := checkstring("总线ID", h.Get("id"), 2, 20)
  if bus.GetBusState(id) != bus.BusStateStop {
    return errors.New("总线已经运行")
  }

  findbus := core.Bus{}
  if err := GetBus(id, h.Ctx(), &findbus); err != nil {
    return err
  }

  tk, err := CreateSchedule(findbus.Timer)
  if err != nil {
    return err
  }

  event := &bus_event{}
  event.init(id, tk, h)
  info, err := bus.NewInfo(findbus.Id, findbus.Type, tk, event)
  if err != nil {
    return err
  }

  sp , err := bus.GetSlotParser(findbus.Type)
  if err != nil {
    return err
  }

  for _, d := range findbus.Datas {
    ds, err := sp.ParseSlot(d.SlotID)
    if err != nil {
      return err
    }
    if err := info.AddData(ds); err != nil {
      return err
    }
    event.push_data(d, ds)
  }

  for _, c := range findbus.Ctrls {
    cs, err := sp.ParseSlot(c.SlotID)
    if err != nil {
      return err
    }
    w , err := _wrap(c.Type, c.Value)
    if err != nil {
      return err
    }
    ctk, err := CreateSchedule(c.Timer)
    if err != nil {
      return err
    }
    if err := info.AddCtrl(cs, ctk, w); err != nil {
      return err
    }
    event.push_ctrl(c, cs, ctk)
  }

  err = bus.StartBus(info)
  if err != nil {
    return err
  }
  event.update(info, h)
  return HttpRet{0, "总线已启动", nil}
}


type w_data_slot struct {
  core.BusSlot
  s bus.Slot
}


type w_ctrl_slot struct {
  core.BusCtrl
  s bus.Slot
  t core.Tick
}


type bus_event struct {
  id        string
  datas     map[string]*w_data_slot
  ctrls     map[string]*w_ctrl_slot
  main_tk   core.Tick
  coll      *mongo.Collection
  for_bus   *mongo.Collection
  ctx       context.Context
}


func (r *bus_event) init(id string, tk core.Tick, h *Ht) {
  r.id = id
  r.datas = make(map[string]*w_data_slot)
  r.ctrls = make(map[string]*w_ctrl_slot)
  r.main_tk = tk
  r.coll = mg.Collection(core.TableBusData)
  r.for_bus = h.Table()
  r.ctx = context.Background()
}


func (r *bus_event) push_data(d core.BusSlot, s bus.Slot) {
  r.datas[d.SlotID] = &w_data_slot{d, s}
}


func (r *bus_event) push_ctrl(c core.BusCtrl, s bus.Slot, t core.Tick) {
  r.ctrls[c.SlotID] = &w_ctrl_slot{c, s, t}
}


func (r *bus_event) update(i *bus.BusInfo, h *Ht) {
  data := bson.M{}
  for _, d := range r.datas {
    data[d.SlotID] = bson.M{ // value/count not change
      "slot_id"   : d.SlotID,
      "slot_desc" : d.SlotDesc,
      "dev_id"    : d.Dev,
      "data_name" : d.Name,
      "data_type" : d.Type.String(),
    }
  }

  ctrl := bson.M{}
  for _, c := range r.ctrls { // count/last_t not change
    ctrl[c.SlotID] = bson.M{
      "slot_id"   : c.SlotID,
      "slot_desc" : c.SlotDesc,
      "dev_id"    : c.Dev,
      "data_name" : c.Name,
      "data_type" : c.Type.String(),

      "value"     : c.Value,
      "start_t"   : c.t.StartTime(),
      "inter_t"   : c.t.Duration(),
      "state"     : "等待发送",
    }
  }

  state := i.State()
  all := bson.M{ // last_t not change
    "_id"     : r.id,
    "state"   : state.String(),
    "start_t" : r.main_tk.StartTime(),
    "inter_t" : r.main_tk.Duration(),
    "data"    : data,
    "ctrl"    : ctrl,
  }

  up := bson.M{ "$set" : all }
  r.update_bstate(h.Ctx(), up)
  r.update_bus(h.Ctx(), state)
}


func (r *bus_event) update_bus(c context.Context, s bus.BusState) {
  up := bson.M{"$set" : bson.M{ "state" : s }}
  filter := bson.M{"_id": r.id}
  if _, err := r.for_bus.UpdateOne(c, filter, up); err != nil {
    log.Println("Update bus fail,", err)
  }
}


func (r *bus_event) update_bstate(c context.Context, up bson.M) {
  filter := bson.M{"_id": r.id}
  opt := options.Update().SetUpsert(true)
  if _, err := r.coll.UpdateOne(c, filter, up, opt); err != nil {
    log.Println("Update bus-state fail,", err)
  }
}


func (r *bus_event) OnStopped() {
  r.update_bus(r.ctx, bus.BusStateStop)
  r.update_bstate(r.ctx, bson.M{ "state": bus.BusStateStop.String() })
}


func (r *bus_event) OnCtrlSended(s bus.Slot, t *time.Time) {
  key := "ctrl."+ s.String()
  up := bson.M{
    "$inc" : bson.M{ key +".count": 1 },
    "$set" : bson.M{ key +".last_t" : t },
  }
  r.update_bstate(r.ctx, up)
}


func (r *bus_event) OnCtrlExit(s bus.Slot) {
  key := "ctrl."+ s.String()
  up := bson.M{
    "$set" : bson.M{ key +".state" : bus.BusStateStop.String() },
  }
  r.update_bstate(r.ctx, up)
}


func (r *bus_event) OnData(s bus.Slot, t *time.Time, d bus.DataWrap) {
  key := "data."+ s.String()
  up := bson.M{
    "$inc" : bson.M{ key +".count": 1 },
    "$set" : bson.M{ 
      "last_t" : t,
      key +".value" : d.String(),
    },
  }
  r.update_bstate(r.ctx, up)
  //TODO: 发送数据到设备数据表
}


func _wrap(t core.DevDataType, v interface{}) (bus.DataWrap, error) {
  switch t {
  case core.DDT_int:
    return &bus.IntData{v.(int)}, nil
  case core.DDT_float:
    return &bus.FloatData{v.(float32)}, nil
  case core.DDT_string:
    return &bus.StringData{v.(string)}, nil
  case core.DDT_sw:
    return &bus.BoolData{v.(bool)}, nil
  case core.DDT_virtual:
    return &bus.StringData{v.(string)}, nil
  }
  return nil, errors.New("无效的类型")
}


func GetBus(id string, ctx context.Context, ret interface{}) error {
  return mg.GetOne(core.TableBus, ctx, id, ret)
}