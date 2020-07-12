package service

import (
	"context"
	"errors"
	"fmt"
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
  aserv(b, ctx, "bus_ctrl_send",  bus_ctrl_send)

  restart_bus()
}


// 系统重启后, 重启正在运行的总线
func restart_bus() {
  log.Println("[[[Restart BUS...")
  filter := bson.M{ "status" : bson.M{"$gt" : bus.BusStateStop} }
  ctx := context.Background()
  cur, err := mg.Collection(core.TableBus).Find(ctx, filter,
      options.Find().SetProjection(bson.M{"_id":1, "status":1, "desc":1}))
  if err != nil {
    log.Println("]]]Restart BUS fail,", err)
    return
  }

  count := 0
  for {
    if !cur.Next(ctx) {
      break
    }
    b := core.Bus{}
    if err = cur.Decode(&b); err != nil {
      log.Println("> Restart BUS fail,", err)
      return
    }

    if b.Status > 0 {
      if err := _bus_start(ctx, b.Id); err != nil {
        log.Println("> Bus start fail", err)
        update_state(ctx, b.Id, bus.BusStateFailStart)
      } else {
        log.Println("> Bus started", b.Id, b.Desc)
        count++
      }
    } else {
      log.Println("> Bus stopped", b.Id, b.Desc)
    }
  }
  if count > 0 {
    log.Println("]]]All Bus restarted.")
  } else {
    log.Println("]]]No Bus need restart.")
  }
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
  uri := h.Get("uri")

  sp, err := bus.GetSlotParser(typ)
  if err != nil {
    return err
  }
  if _, err := sp.ParseURI(uri); err != nil {
    return err
  }
  if !hasTimer(tm) {
    return errors.New("定时器不存在 "+ tm)
  }

  d := bson.M{
    "_id"       : id,
    "uri"       : uri,
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
  uri := h.Get("uri")

  if bus.GetBusState(id) != bus.BusStateStop {
    return errors.New("总线运行中, 禁止修改 "+ id)
  }
  if !hasTimer(tm) {
    return errors.New("定时器不存在 "+ tm)
  }

  findbus := core.Bus{}
  if err := GetBus(id, h.Ctx(), &findbus); err != nil {
    return err
  }
  sp , err := bus.GetSlotParser(findbus.Type)
  if err != nil {
    return err
  }
  if _, err := sp.ParseURI(uri); err != nil {
    return err
  }

  d := bson.M{
    "uri"       : uri,
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

    value_str := h.Get("value")
    value, err := pd.Type.Parse(value_str)
    if err != nil {
      return HttpRet{5, "无法解析参数", err}
    }

    slot_conf["timer"] = timer_id
    slot_conf["value"] = value
    set[_ctrl_slot_key(slot_id, value_str)] = slot_conf
  }
  return h.Crud().Update(id, bson.M{ "$set": set })
}


func _ctrl_slot_key(slotid, value string) string {
  return fmt.Sprintf("ctrl_slot.%s+%s", slotid, value)
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
    // key = "ctrl_slot."+ slot_id
    key = _ctrl_slot_key(slot_id, h.Get("value"))
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
    return HttpRet{3, "数据不存在", key}
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
  ret, err := c.DRead(id)
  if err != nil {
    return HttpRet{3, "没有总线实时数据", err.Error()}
  }
  b , err := bus.GetBus(id)
  if err == nil {
    ret["logs"] = b.GetLog()
  }
  return HttpRet{0, c.info, ret}
}


func bus_ctrl_send(h *Ht) interface{} {
  id := checkstring("总线ID", h.Get("id"), 2, 20)
  slot_id := checkstring("控制槽ID", h.Get("slot_id"), 1, 99)
  v := checkstring("发送值", h.Get("value"), 1, 99)

  inf, err := bus.GetBus(id)
  if err != nil {
    return err
  }
  slot, err := inf.ParseSlot(slot_id)
  if err != nil {
    return err
  }
  if slot.Type() != bus.SlotCtrl {
    return errors.New("不是控制槽")
  }
  
  value := bus.StringData{ D: v }
  if err := inf.SendCtrl(slot, &value); err != nil {
    return err
  }
  return HttpRet{0, "控制已发送", slot_id}
}


func bus_stop(h *Ht) interface{} {
  id := checkstring("总线ID", h.Get("id"), 2, 20)
  if err := bus.StopBus(id); err != nil {
    return err
  }
  update_state(h.Ctx(), id, bus.BusStateStop)
  return HttpRet{0, "总线已终止", nil}
}


func bus_start(h *Ht) interface{} {
  id := checkstring("总线ID", h.Get("id"), 2, 20)
  if bus.GetBusState(id) != bus.BusStateStop {
    return errors.New("总线已经运行")
  }
  if err := _bus_start(h.Ctx(), id); err != nil {
    update_state(h.Ctx(), id, bus.BusStateFailStart)
    return err;
  }
  return HttpRet{0, "总线已启动", nil}
}


func _bus_start(ctx context.Context, id string) error {
  findbus := core.Bus{}
  if err := GetBus(id, ctx, &findbus); err != nil {
    return err
  }

  tk, err := CreateSchedule(findbus.Timer)
  if err != nil {
    return err
  }

  event := &bus_event{}
  info, err := bus.NewInfo(findbus.Uri, findbus.Id, findbus.Type, tk, event)
  if err != nil {
    return err
  }
  event.init(id, tk, info)

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
    if err := event.init_dev_script(d.Dev); err != nil {
      return &HttpRet{6, "加载设备脚本时出错", err.Error()}
    }
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
  event.update(ctx, info)
  return nil
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
  info      *bus.BusInfo
  devjs     map[string]*ScriptRuntime
}


func (r *bus_event) init(id string, tk core.Tick, i *bus.BusInfo) {
  r.id = id
  r.datas = make(map[string]*w_data_slot)
  r.ctrls = make(map[string]*w_ctrl_slot)
  r.main_tk = tk
  r.ctx = context.Background()
  r.info = i
  r.devjs = make(map[string]*ScriptRuntime)
}


// 插入一个数据槽
func (r *bus_event) push_data(d core.BusSlot, s bus.Slot) {
  r.datas[d.SlotID] = &w_data_slot{d, s}
}


// 插入一个控制槽
func (r *bus_event) push_ctrl(c core.BusCtrl, s bus.Slot, t core.Tick) {
  r.ctrls[c.SlotID] = &w_ctrl_slot{c, s, t}
}


// 初始化设备脚本
func (r *bus_event) init_dev_script(devid string) error {
  if _, has := r.devjs[devid]; has {
    return nil
  }

  dev := core.Device{}
  if err := GetDevice(r.ctx, devid, &dev); err != nil {
    return err
  }
  devp := core.DevProto{}
  if err := GetDevProto(r.ctx, dev.ProtoId, &devp); err != nil {
    return err
  }
  js := core.DevScript{}
  if err := GetDevScript(r.ctx, devp.Script, &js); err != nil {
    return err
  }
  sr, err := BuildDevScript(js.Id, js.Js)
  if err != nil {
    return err
  }
  r.devjs[devid] = sr
  return nil
}


//
// 更新数据库状态: 总线状态, 总线实时数据; 增加设备引用计数.
//
func (r *bus_event) update(ctx context.Context, i *bus.BusInfo) {
  data := bson.M{}
  for _, d := range r.datas {
    data[d.SlotID] = bson.M{ // value/count not change
      "slot_id"   : d.SlotID,
      "slot_desc" : d.SlotDesc,
      "dev_id"    : d.Dev,
      "data_name" : d.Name,
      "data_type" : d.Type.String(),
      "data_desc" : d.Desc,
    }
    device_ref.Add(d.Dev)
  }

  ctrl := bson.M{}
  for _, c := range r.ctrls { // count/last_t not change
    ctrl[c.SlotID] = bson.M{
      "slot_id"   : c.SlotID,
      "slot_desc" : c.SlotDesc,
      "dev_id"    : c.Dev,
      "data_name" : c.Name,
      "data_type" : c.Type.String(),
      "data_desc" : c.Desc,

      "value"     : c.Value,
      "start_t"   : c.t.StartTime(),
      "inter_t"   : c.t.Duration(),
      "status"    : "等待发送",
    }
    device_ref.Add(c.Dev)
  }

  state := i.State()
  all := bson.M{ // last_t not change
    "_id"     : r.id,
    "status"  : state.String(),
    "start_t" : r.main_tk.StartTime(),
    "inter_t" : r.main_tk.Duration(),
    "data"    : data,
    "ctrl"    : ctrl,
  }

  up := bson.M{ "$set" : all }
  update_bus_ldata(ctx, r.id, up)
  update_bus(ctx, r.id, state)
}


func update_bus(c context.Context, id string, s bus.BusState) {
  up := bson.M{"$set" : bson.M{ "status" : s }}
  filter := bson.M{"_id": id}
  coll := mg.Collection(core.TableBus)
  if _, err := coll.UpdateOne(c, filter, up); err != nil {
    log.Println("Update bus fail,", err)
  }
}


func update_bus_ldata(c context.Context, id string, up bson.M) {
  filter := bson.M{"_id": id}
  opt := options.Update().SetUpsert(true)
  coll := mg.Collection(core.TableBusData)
  if _, err := coll.UpdateOne(c, filter, up, opt); err != nil {
    log.Println("Update bus-state fail,", err)
  }
}


func update_state(ctx context.Context, id string, s bus.BusState) {
  update_bus_ldata(ctx, id, bson.M{"status": s.String()})
  update_bus(ctx, id, s)
}


func (r *bus_event) OnStopped() {
  update_bus(r.ctx, r.id, bus.BusStateStop)
  update_bus_ldata(r.ctx, r.id, bson.M{ "status": bus.BusStateStop.String() })

  for _, d := range r.datas {
    device_ref.Free(d.Dev)
  }

  for _, c := range r.ctrls {
    device_ref.Free(c.Dev)
  }
}


func (r *bus_event) OnCtrlSended(s bus.Slot, t *time.Time) {
  key := "ctrl."+ s.String()
  up := bson.M{
    "$inc" : bson.M{ key +".count": 1 },
    "$set" : bson.M{ key +".last_t" : t },
  }
  update_bus_ldata(r.ctx, r.id, up)
}


func (r *bus_event) OnCtrlExit(s bus.Slot) {
  key := "ctrl."+ s.String()
  up := bson.M{
    "$set" : bson.M{ key +".status" : bus.BusStateStop.String() },
  }
  update_bus_ldata(r.ctx, r.id, up)
}


func (r *bus_event) OnData(s bus.Slot, t *time.Time, d bus.DataWrap) {
  info, has := r.datas[s.String()]
  if !has {
    r.info.Log("系统错误, 在未配置的数据槽上发送数据 "+ s.String())
  }
  // 脚本过滤参数
  if devjs, has := r.devjs[info.Dev]; has {
    var err error
    d, err = devjs.OnData(&info.BusSlot, t, d)
    if err != nil {
      r.info.Log("设备", info.Dev, "脚本", devjs.Name, "错误", err)
    }
  }

  // 发送数据到总线实时数据表
  key := "data."+ s.String()
  up := bson.M{
    "$inc" : bson.M{ key +".count": 1 },
    "$set" : bson.M{ 
      "last_t" : t,
      key +".value" : d.String(),
    },
  }
  update_bus_ldata(r.ctx, r.id, up)
  
  // 发送数据到设备数据表
  err := send_dev_data(r.ctx, &info.BusSlot, d, t)
  if err != nil {
    r.info.Log("保存数据错误, "+ err.Error())
  }
  
  err = UpdateDataCount(r.ctx, info.Dev, t)
  if err != nil {
    r.info.Log("更新设备状态错误, "+ err.Error())
  }
}


func _wrap(t core.DevDataType, v interface{}) (bus.DataWrap, error) {
  switch t {
  case core.DDT_int:
    return &bus.Int64Data{ v.(int64) }, nil
  case core.DDT_float:
    return &bus.Float64Data{ v.(float64) }, nil
  case core.DDT_string:
    return &bus.StringData{ v.(string) }, nil
  case core.DDT_sw:
    return &bus.BoolData{ v.(bool) }, nil
  case core.DDT_virtual:
    return &bus.StringData{ v.(string) }, nil
  }
  return nil, errors.New("无效的类型")
}


func GetBus(id string, ctx context.Context, ret interface{}) error {
  return mg.GetOne(core.TableBus, ctx, id, ret)
}