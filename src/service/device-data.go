package service

import (
	"context"
	"errors"
	"fmt"
	"ic1101/brick"
	"ic1101/src/bus"
	"ic1101/src/core"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func installDeviceDataService(b *brick.Brick) {
  mg.CreateIndex(core.TableDevice, &bson.D{{"dev", 1}})
  ctx := &ServiceGroupContext{core.TableDevData, "设备数据"}
  aserv(b, ctx, "dev_data_read", dev_data_read)
  dserv(b, ctx, "dev_data_range", dev_data_range)
}


func dev_data_read(h *Ht) interface{} {
  rg := h.GetInt("range", -1)
  if rg < 0 {
    return errors.New("无效的 range 参数")
  }

  did  := checkstring("设备ID", h.Get("did"), 2, 20)
  name := checkstring("数据名称", h.Get("name"), 2, 20)
  tr   := core.TimeRange(rg)
  // time 参数是 0 时区偏移的时间字符串
  tm, err := time.Parse(time.RFC3339, h.Get("time"))
  if err != nil {
    return err
  }
  // 数据表中的 id 是已经转换为本地时间的字符串
  tm = tm.Local()
  
  id := tr.GetId(did, name, &tm)
  return h.Crud().Read(id)
}


func dev_data_range(h *Ht) interface{} {
  return HttpRet{ 0, "map", core.TimeRangeMap }
}


//
// 删除设备的所有数据
//
func delete_dev_data(ctx context.Context, devid string) error {
  table := mg.Collection(core.TableDevData)
  _, err := table.DeleteMany(ctx, bson.M{"dev": devid})
  return err
}


//
// 发送设备数据
//
func send_dev_data(ctx context.Context, i *core.BusSlot, 
                   data bus.DataWrap, t *time.Time) (ret error) {
  v, err := get_saved_data(i.Type, data)
  if err != nil {
    return err
  }

  up := _update_dd{
    ctx     : ctx,
    table   : mg.Collection(core.TableDevData),
    filter  : bson.M{},
    up      : bson.M{"dev" : i.Dev},
    opt     : options.Update().SetUpsert(true),
    val     : v,
    set     : bson.M{},
  }
  up.set["$set"] = up.up

  defer func() {
    if e := recover(); e != nil {
      ret = e.(error)
    }
  }()

  up.update(core.GetDDYearID(i.Dev, i.Name, t),   
            fmt.Sprintf("v.%d", t.Year()) )
  up.update(core.GetDDMonthID(i.Dev, i.Name, t),  
            fmt.Sprintf("v.%d", t.Month()) )
  up.update(core.GetDDDayID(i.Dev, i.Name, t),    
            fmt.Sprintf("v.%d", t.Day()) )

  up.update(core.GetDDHourID(i.Dev, i.Name, t),   
            fmt.Sprintf("v.%d", t.Hour()) )
  up.update(core.GetDDMinuteID(i.Dev, i.Name, t), 
            fmt.Sprintf("v.%d", t.Minute()) )
  up.update(core.GetDDSecondID(i.Dev, i.Name, t), 
            fmt.Sprintf("v.%d", t.Second()) )
  return nil
}


func get_saved_data(t core.DevDataType, data bus.DataWrap) (interface{}, error) {
  switch t {
  case core.DDT_int:
    return data.Int(), nil
  case core.DDT_float:
    return data.Float(), nil
  case core.DDT_sw:
    return data.Bool(), nil
  case core.DDT_string:
    return data.String(), nil
  case core.DDT_virtual: //TODO: 用脚本处理虚拟数据
    return data.String(), nil
  default:
    return nil, errors.New("未知的数据类型")
  }
}


type _update_dd struct {
  ctx     context.Context
  table   *mongo.Collection
  filter  bson.M
  up      bson.M
  opt     *options.UpdateOptions
  val     interface{}
  set     bson.M
}


func (u *_update_dd) update(id, vKey string) {
  u.filter["id"] = id
  u.up["_id"] = id
  u.up["l"] = u.val
  u.up[vKey] = u.val
  defer delete(u.up, vKey)
  
  _, err := u.table.UpdateOne(u.ctx, u.filter, u.set, u.opt)
  if err != nil {
    panic(err)
  }
}