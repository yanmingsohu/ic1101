package service

import (
	"errors"
	"fmt"
	"ic1101/brick"
	"ic1101/src/core"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func installTimerService(b *brick.Brick) {
  mg.CreateIndex(core.TableDevice, &bson.D{{"_id", "text"}, {"desc", "text"}})
  ctx := &ServiceGroupContext{core.TableTimer, "定时器"}

  aserv(b, ctx, "timer_count",  timer_count)
  aserv(b, ctx, "timer_list",   timer_list)
  aserv(b, ctx, "timer_create", timer_create)
  aserv(b, ctx, "timer_delete", timer_delete)
  aserv(b, ctx, "timer_update", timer_update)

  aserv(b, ctx, "timer_test", timer_test) // 测试用, 注释掉
}


func timer_count(h *Ht) interface{} {
  return h.Crud().PageInfo()
}


func timer_list(h *Ht) interface{} {
  return h.Crud().List(func(opt *options.FindOptions) {
    opt.SetProjection(bson.M{
      "desc":1, "duration":1, "loop":1, "md":1, "cd":1, "delay":1 })
  })
}


func timer_create(h *Ht) interface{} {
  id   := checkstring("定时器ID", h.Get("id"), 2, 20)
  dur, err := time.ParseDuration("1s")
  if err != nil {
    return err
  }

  delay := core.TimerDelay{-1, -1, -1, -1, -1}
  d := bson.M{
    "_id"       : id,
    "desc"      : h.Get("desc"),
    "duration"  : dur,
    "loop"      : h.GetBool("$loop"),
    "cd"        : time.Now(),
    "md"        : "",
    "delay"     : delay,
  }
  return h.Crud().Create(d)
}


func timer_delete(h *Ht) interface{} {
  id := checkstring("定时器ID", h.Get("id"), 2, 20)
  return h.Crud().Delete(id)
}


func timer_update(h *Ht) interface{} {
  id := checkstring("定时器ID", h.Get("id"), 2, 20)
  dur, err := time.ParseDuration(h.Get("duration"))
  if err != nil {
    return err
  }

  delay := core.TimerDelay{
    Mon  : h.GetInt("d.mon", -1),
    Day  : h.GetInt("d.day", -1),
    Hour : h.GetInt("d.hour", -1),
    Min  : h.GetInt("d.min", -1),
    Sec  : h.GetInt("d.sec", -1),
  }

  d := bson.M{
    "desc"      : h.Get("desc"),
    "duration"  : dur,
    "loop"      : h.GetBool("$loop"),
    "md"        : time.Now(),
    "delay"     : delay,
  }
  return h.Crud().Update(id, bson.D{{"$set", d}});
}


// 测试: 启动一个不能停止的任务
func timer_test(h *Ht) interface{} {
  id := checkstring("定时器ID", h.Get("id"), 2, 20)
  tk, err := CreateSchedule(id)
  if err != nil {
    return err
  }

  c := 0
  tk.Start(func() {
    c++
    log.Print("Timer ", id, " Ticker ",  c)
  }, func() {
    log.Print("Timer ", id, " Stoped")
  })
  log.Print(tk)
  return HttpRet{0, "测试线程已经启动", tk}
}


//
// core.Tick 接口的实现
//
type _Tick struct {
  d       *core.Timer
  tm      *time.Timer
  tk      *time.Ticker
  BeginAt *time.Time
  running bool
  onStop  func()
  mutex   *sync.RWMutex
}


func _NewTick(t *core.Timer) core.Tick {
  return &_Tick{t, nil, nil, nil, false, nil, new(sync.RWMutex)}
}


func (t *_Tick) Start(task func(), on_stop func()) {
  t.mutex.Lock()
  defer t.mutex.Unlock()

  if t.running {
    panic(errors.New("定时器已经启动"))
  }
  t.running = true
  t.onStop = on_stop

  n := time.Now()
  year := n.Year()
  // 基于当前时间计算启动时钟, 忽略的时钟部分由当前时间代替
  mon  := _cho_time(t.d.Delay.Mon, int(n.Month()))
  day  := _cho_time(t.d.Delay.Day, n.Day())
  hour := _cho_time(t.d.Delay.Hour, n.Hour())
  min  := _cho_time(t.d.Delay.Min, n.Minute())
  sec  := _cho_time(t.d.Delay.Sec, n.Second())

  then := time.Date(year, time.Month(mon), day, hour, min, sec, 0, time.Local)
  then = AddMinimumUnit(n, then)
  t.BeginAt = &then

  t.tm = time.AfterFunc(then.Sub(n), func() {
    t.tk = time.NewTicker(t.d.Duration)
    task()
    go t.runTask(task)
  })
}


func (t *_Tick) runTask(task func()) {
  if t.d.Loop {
    for range t.tk.C {
      task()
    }
  }
  t.Stop()
}


func (t *_Tick) Stop() {
  t.mutex.Lock()
  defer t.mutex.Unlock()

  if !t.running {
    panic(errors.New("定时器已经停止"))
  }
  t.running = false

  if t.tk != nil {
    t.tk.Stop()
  }

  if t.tm != nil {
    t.tm.Stop()
  }

  if t.onStop != nil {
    t.onStop()
  }
}


func (t *_Tick) IsRunning() bool {
  t.mutex.RLock()
  defer t.mutex.RUnlock()
  return t.running
}


func (t *_Tick) String() string {
  t.mutex.RLock()
  defer t.mutex.RUnlock()
  return fmt.Sprintf("Timer [%s] start on %s Per %s", 
    t.d.Id, t.BeginAt, t.d.Duration)
}


//
// 保证 then 在 n 的时间之后, 依次增加 秒/分/时 等, 并返回时间
//
func AddMinimumUnit(n, then time.Time) time.Time {
  if then.Before(n) { // 阶梯加时
    test := then.Add(time.Second)
    if !test.Before(n) {
      return test
    }
    test = then.Add(time.Minute)
    if !test.Before(n) {
      return test
    }
    test = then.Add(time.Hour)
    if !test.Before(n) {
      return test
    }
    test = then.AddDate(0, 0, 1)
    if !test.Before(n) {
      return test
    }
    test = then.AddDate(0, 1, 0)
    if !test.Before(n) {
      return test
    }
    return then.AddDate(1, 0, 0)
  }
  return then
}


func _cho_time(a, b int) int {
  if a < 0 {
    return b
  }
  return a
}


//
// 创建一个计时器对象, 用于运行任务
//
func CreateSchedule(id string) (core.Tick, error) {
  filter := bson.M{ "_id" : id }
  table := mg.Collection(core.TableTimer)
  t := core.Timer{}

  if err := table.FindOne(nil, filter).Decode(&t); err != nil {
    return nil, err
  }
  return _NewTick(&t), nil
}