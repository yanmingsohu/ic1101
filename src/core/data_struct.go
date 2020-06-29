package core

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

/*
Table: login_user {
  _id(string)  			: 用户名, 不重复
  pass(string) 			: 两次加密密钥
  role(string) 			: 用户角色
  isroot(bool) 			: 超级用户
  wexin(string) 		: 微信号
  tel(string) 			: 电话号
  email(string) 		: 邮箱
  regdata(string) 	: 注册时间
  logindata(string) : 最后登录时间
}
*/
type LoginUser struct {
  Name string `yaml:"username" bson:"_id"`
  Pass string `yaml:"password" bson:"pass"`
  Role string `bson:"role"`
  IsRoot bool `bson:"isroot"`
  Auths map[string]bool 
}

const TableLoginUser = "login_user"


/*
Table: dict {
  _id(string)       : 字典(ID)名称
  desc(string)			: 说明
  content           : 字典内容
  cd(time)          : 创建时间
  md(time)  				: 修改时间
}
dict: content {
  key : value
}
*/
type Dict struct {
  Id 			string  					`bson:"_id"`
  Desc 		string 						`bson:"desc"`
  Cd      time.Time         `bson:"cd"`
  Md      time.Time         `bson:"md"`
  Content map[string]string `bson:"content"`
}

const TableDict = "dict"


/*
Table: role {
  _id(string)     : 角色名
  desc(string)    : 说明
  cd(time)        : 创建时间
  md(time)        : 修改时间
  rules([]string) : 权限列表
}
*/
type Role struct {
  Id      string `bson:"_id"`
}

const TableRule = "role"


/*
Table: dev-proto {
  _id(string)    : 原型id
  desc(string)   : 说明
  changeid       : 修改次数flag
  cd(time)       : 创建时间
  md(time)       : 修改时间
  script(string) : '脚本名, 用于处理虚拟数据, 有接口标准'

  attrs : 属性信息列表 [
    { name     : '属性名'
      desc     : '说明'
      type     : DevAttrType '类型索引'
      notnull  : bool 不能空 
      defval   : '默认值'
      dict     : '只在字典类型时有效'
      max      : 最大值
      min      : 最小值
    }
  ]

  datas : 输入数据列表 [
    { name     : '数据名'
      desc     : '说明'
      type     : DevDataType '数据类型'
    }
  ]

  ctrls : 控制列表 [
    { name    : '名称'
      desc    : '说明'
      type    : DevDataType '发送的控制数据'
    }
  ]
}
*/
type DevProto struct {
  Id        string     `bson:"_id"`
  Desc      string     `bson:"desc"`
  ChangeId  int        `bson:"changeid"`
  Cd        time.Time  `bson:"cd"`
  Md        time.Time  `bson:"md"`
  Script    string     `bson:"script"`

  Attrs []DevProtoAttr `bson:"attrs"`
  Datas []DevProtoData `bson:"datas"`
  Ctrls []DevProtoData `bson:"ctrls"`
}

type DevProtoAttr struct {
  Name    string      `bson:"name"`
  Desc    string      `bson:"desc"`
  Type    DevAttrType `bson:"type"`
  Notnull bool        `bson:"notnull"`
  Defval  string      `bson:"defval"`
  Dict    string      `bson:"dict"`
  Max     int64       `bson:"max"`
  Min     int64       `bson:"min"`
}

type DevProtoData struct {
  Name    string      `bson:"name"`
  Desc    string      `bson:"desc"`
  Type    DevDataType `bson:"type"`
}

const TableDevProto = "dev-proto"

type DevAttrType int
type DevDataType int

const (
  DDT_int       DevDataType = 1 // 整数类型
  DDT_float     DevDataType = 2 // 浮点类型
  DDT_virtual   DevDataType = 3 // 虚拟数据
  DDT_sw        DevDataType = 4 // 开关类型
  DDT_string    DevDataType = 5

  DAT_string    DevAttrType = 100 // 字符串
  DAT_number    DevAttrType = 101 // 数字
  DAT_dict      DevAttrType = 102 // 字典
  DAT_date      DevAttrType = 103
)

var DDT__map = map[DevDataType]string {
  DDT_int       : "整数",
  DDT_float     : "浮点",
  DDT_sw        : "开关",
  DDT_string    : "字符串",
  DDT_virtual   : "虚拟",
}

var DAT__map = map[DevAttrType]string {
  DAT_string    : "字符串",
  DAT_number    : "数字",
  DAT_dict      : "字典",
  DAT_date      : "日期",
}


/*
Table: device {
  _id       : "设备ID"
  desc      : "说明"
  tid       : "原型id"
  changeid  : 引用原型的修改次数flag, 当设备小于原型则属性需要同步
  md(time)  : 修改时间
  cd(time)  : 创建时间
  
  dd(time)  : 最后数据时间
  dc(int64) : 数据量

  attrs : 属性值 {
    "属性名, 与原型对应" : string 属性值
  }
}
*/
type Device struct {
  Id string `bson:"_id"`
}

const TableDevice = "device"


/*
设备数据表设计:

每个设备有自己的数据表, 表名: `data@[device-id]`
每个时间粒度有一个单独文档, 文档名称/内容:

  所有年份数据:  {
    _id : year$[data-name] (数据名)
    l : 最后插入的数据
    v : (数据 map) {
      Y : n年数据, (数字类型)
      ...
    }
  }

  当年所有月数据: {
    _id : month$[data-name]@Y
    l : 最后插入的数据
    v : {
      1 : 1月数据 (月份是数字类型)
      ...
      12 : 12月数据
    }
  }

  日数据: {
    _id : day$[data-name]@Y-M
    l : 最后插入的数据
    v : {
      1 : x月1日数据
      ...
      31 : 31日数据
    }
  }

  小时数据: {
    _id : hour$[data-name]@Y-M-D
    l : 最后插入的数据
    v : {
      0 : 0点数据
      ...
      23 : 23点数据
    }
  }

  分钟数据:  {
    _id : minute$[data-name]@Y-M-D_h
    l : 最后插入的数据
    v : {
      0 : 0分数据
      ...
      59: 59分数据
    }
  }

  秒数据:  {
    _id : second$[data-name]@Y-M-D_h:m
    l : 最后插入的数据
    v : {
      0 : 0秒数据
      ...
      59: 59秒数据
    }
  }

** 所有时间数据都没有用 0 补位
** fmt.Print(t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(), t.Second())
*/
type DeviceData struct {
  Id string `bson:"_id"`
  V  bson.M `bson:"v"`
}

type TimeRange int

const (
  TimeRangeYear   TimeRange = 1
  TimeRangeMonth  TimeRange = 2
  TimeRangeDay    TimeRange = 3
  TimeRangeHour   TimeRange = 4
  TimeRangeMinute TimeRange = 5
  TimeRangeSecond TimeRange = 6
)

var TimeRangeMap = map[TimeRange]string {
  TimeRangeYear   : "年",
  TimeRangeMonth  : "月",
  TimeRangeDay    : "日",
  TimeRangeHour   : "时",
  TimeRangeMinute : "分",
  TimeRangeSecond : "秒",
}

//
// 返回数据表名
//
func TableDevData(devid string) string {
  return "data@"+ devid
}

//
// 返回数据的 id
//
func DevDataID(r TimeRange, dataName string, t *time.Time) string {
  switch (r) {
  case TimeRangeYear:
    return "year$"+ dataName

  case TimeRangeMonth:
    return fmt.Sprintf("month$%s@%d", dataName, t.Year())

  case TimeRangeDay:
    return fmt.Sprintf("day$%s@%d-%d", dataName, t.Year(), t.Month())

  case TimeRangeHour:
    return fmt.Sprintf("hour$%s@%d-%d-%d", dataName, 
            t.Year(), t.Month(), t.Day())

  case TimeRangeMinute:
    return fmt.Sprintf("minute$%s@%d-%d-%d_%d", dataName, 
            t.Year(), t.Month(), t.Day(), t.Minute())

  case TimeRangeSecond:
    return fmt.Sprintf("minute$%s@%d-%d-%d_%d:%d", dataName, 
            t.Year(), t.Month(), t.Day(), t.Minute(), t.Second())
            
  default:
    panic("无效的时间范围")
  }
}


/*
Table: bus {
  _id(string)    : 总线id
  desc(string)   : 说明
  timer          : 定时器id
  cd(time)       : 创建时间
  md(time)       : 修改时间
}
*/
type Bus struct {
  Id string `bson:"_id"`
}

const TableBus = "bus"


/*
Table: timer {
  _id(string)      : 定时器 id
  desc(string)     : 说明
  duration         : 时间间隔
  loop(bool)       : 重复执行/执行一次
  cd(time)         : 创建时间
  md(time)         : 修改时间

  delay : 启动时间, -1表示忽略 {
    mon : 月, day : 日, hour: 时, min: 分, sec: 秒
  }
}
*/
type Timer struct {
  Id        string        `bson:"_id"`
  Desc      string        `bson:"desc"`
  Duration  time.Duration `bson:"duration"`
  Loop      bool          `bson:"loop"`
  // 当时钟相符时才启动任务, 
  // 并且基于当前时间以最小单位向后推进时钟
  Delay     TimerDelay    `bson:"delay"`
}

type TimerDelay struct {
  Mon  int `bson:"mon"`
  Day  int `bson:"day"`
  Hour int `bson:"hour"`
  Min  int `bson:"min"`
  Sec  int `bson:"sec"`
}

const TableTimer = "timer"

//
// 定时器接口
//
type Tick interface {
  // 该方法只能调用一次S
  Start(f func())
  Stop()
  IsRunning() bool
}