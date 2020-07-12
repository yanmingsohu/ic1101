package core

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"gopkg.in/yaml.v2"
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

//
// 在列表中寻找匹配的数据名称并返回该对象
//
func FindProtoDataByName(a []DevProtoData, name string) (*DevProtoData, error) {
  for _, d := range a {
    if d.Name == name {
      return &d, nil
    }
  }
  return nil, errors.New("在原型中找不到数据/控制槽: "+ name)
}

//
// 按数据类型转换数据并返回
//
func (t DevDataType) Parse(s string) (interface{}, error) {
  switch t {
  case DDT_int:
    return strconv.ParseInt(s, 10, 32)
  case DDT_float:
    return strconv.ParseFloat(s, 32)
  case DDT_string:
    return s, nil
  case DDT_sw:
    return strconv.ParseBool(s)
  case DDT_virtual:
    return s, nil
  }
  return nil, errors.New("无效的类型")
}

func (t DevDataType) String() string {
  return DDT__map[t]
}

func (t DevAttrType) String() string {
  return DAT__map[t]
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
  Id      string `bson:"_id"`
  ProtoId string `bson:"tid"`
}

const TableDevice = "device"


/*
设备数据表设计:

每个设备有自己的数据表, 表名: `dev-data`
每个时间粒度有一个单独文档, 文档名称/内容:

  所有年份数据:  {
    _id : !yr~[device-id]$[data-name] (数据名)
    dev : 设备ID - 需要索引
    l : 最后插入的数据
    v : (数据 map) {
      Y : n年数据, (数字类型)
      ...
    }
  }

  当年所有月数据: {
    _id : !mo~[device-id]$[data-name]@Y
    dev : 设备ID
    l : 最后插入的数据
    v : {
      1 : 1月数据 (月份是数字类型)
      ...
      12 : 12月数据
    }
  }

  日数据: {
    _id : !dy~[device-id]$[data-name]@Y-M
    dev : 设备ID
    l : 最后插入的数据
    v : {
      1 : x月1日数据
      ...
      31 : 31日数据
    }
  }

  小时数据: {
    _id : !hr~[device-id]$[data-name]@Y-M-D
    dev : 设备ID
    l : 最后插入的数据
    v : {
      0 : 0点数据
      ...
      23 : 23点数据
    }
  }

  分钟数据:  {
    _id : !mi~[device-id]$[data-name]@Y-M-D_h
    dev : 设备ID
    l : 最后插入的数据
    v : {
      0 : 0分数据
      ...
      59: 59分数据
    }
  }

  秒数据:  {
    _id : !se~[device-id]$[data-name]@Y-M-D_h:m
    dev : 设备ID
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

const TableDevData = "dev-data"

type TimeRange int

const (
  TimeRangeAllYear TimeRange = iota
  TimeRangeYear    TimeRange = iota
  TimeRangeMonth   TimeRange = iota
  TimeRangeDay     TimeRange = iota
  TimeRangeHour    TimeRange = iota
  TimeRangeMinute  TimeRange = iota
  TimeRangeSecond  TimeRange = iota
)

var TimeRangeMap = map[TimeRange]string {
  TimeRangeYear    : "年",
  TimeRangeMonth   : "月",
  TimeRangeDay     : "日",
  TimeRangeHour    : "时",
  TimeRangeMinute  : "分",
  TimeRangeSecond  : "秒",
}


//
// 返回数据表的 id, (设备id, 数据名称, 时间)
//
func (r TimeRange) GetId(did, name string, t *time.Time) string {
  switch (r) {
  case TimeRangeYear:
    return GetDDYearID(did, name, t)

  case TimeRangeMonth:
    return GetDDMonthID(did, name, t)

  case TimeRangeDay:
    return GetDDDayID(did, name, t)

  case TimeRangeHour:
    return GetDDHourID(did, name, t)

  case TimeRangeMinute:
    return GetDDMinuteID(did, name, t)

  case TimeRangeSecond:
    return GetDDSecondID(did, name, t)
            
  default:
    panic("无效的时间范围")
  }
}


func GetDDYearID(did, name string, t *time.Time) string {
  return fmt.Sprintf("!yr~%s$%s", did, name)
}


func GetDDMonthID(did, name string, t *time.Time) string {
  return fmt.Sprintf("!mo~%s$%s@%d", did, name, t.Year())
}


func GetDDDayID(did, name string, t *time.Time) string {
  return fmt.Sprintf("!dy~%s$%s@%d-%d", did, name, t.Year(), t.Month())
}


func GetDDHourID(did, name string, t *time.Time) string {
  return fmt.Sprintf("!hr~%s$%s@%d-%d-%d", 
          did, name, t.Year(), t.Month(), t.Day())
}


func GetDDMinuteID(did, name string, t *time.Time) string {
  return fmt.Sprintf("!mi~%s$%s@%d-%d-%d_%d", 
          did, name, t.Year(), t.Month(), t.Day(), t.Hour())
}


func GetDDSecondID(did, name string, t *time.Time) string {
  return fmt.Sprintf("!se~%s$%s@%d-%d-%d_%d:%d", did, name, 
          t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())
}


/*
Table: bus {
  _id(string)    : 总线id
  uri            : 服务器启动参数
  desc(string)   : 说明
  timer          : 定时器id
  cd(time)       : 创建时间
  md(time)       : 修改时间
  type           : 总线类型(不可改)
  status(int)    : 状态

  data_slot : {
    "slot_id" : { 
      slot_id,
      slot_desc
      dev_id, 
      data_name, 
      data_type,
      data_desc,
    }
  }

  ctrl_slot : {
    "slot_id" : {
      slot_id,
      slot_desc
      dev_id, 
      data_name, 
      data_type,
      data_desc,

      timer : 控制定时器, 定时发数
      value : 发送的数据
    }
  }
}
*/
type Bus struct {
  Id      string             `bson:"_id"`
  Uri     string             `bson:"uri"`
  Desc    string             `bson:"desc"`
  Timer   string             `bson:"timer"`
  Type    string             `bson:"type"`
  Status  int                `bson:"status"`
  Datas   map[string]BusSlot `bson:"data_slot"`
  Ctrls   map[string]BusCtrl `bson:"ctrl_slot"`
}

type BusSlot struct {
  SlotID    string      `bson:"slot_id"   json:"slot_id"`
  SlotDesc  string      `bson:"slot_desc" json:"slot_desc"`
  Dev       string      `bson:"dev_id"    json:"dev_id"`
  Name      string      `bson:"data_name" json:"data_name"`
  Desc      string      `bson:"data_desc" json:"data_desc"`
  Type      DevDataType `bson:"data_type" json:"data_type"`
}

type BusCtrl struct {
  SlotID    string      `bson:"slot_id"   json:"slot_id"`
  SlotDesc  string      `bson:"slot_desc" json:"slot_desc"`
  Dev       string      `bson:"dev_id"    json:"dev_id"`
  Name      string      `bson:"data_name" json:"data_name"` //难以纠正的错误
  Desc      string      `bson:"data_desc" json:"data_desc"` //..
  Type      DevDataType `bson:"data_type" json:"data_type"` //..
  Timer     string      `bson:"timer"     json:"timer"`
  Value     interface{} `bson:"value"     json:"value"`
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
  // 该方法只能调用一次, 任务启动时调用 task(), 任务终止时调用 onStop()
  // 多次调用该方法将抛出异常
  Start(task func(), onStop func())
  // 终止任务, onStop 会被调用
  Stop()
  IsRunning() bool
  // 返回间隔时间
  Duration() time.Duration
  // 返回首次启动时间, 该方法必须在启动后调用, 否则返回 nil
  StartTime() *time.Time
}


/*
总线最后状态表, 实时变更
  * 在启动/退出时改变
  ** 每次数据改变

Table: bus_ldata {
  _id(string) : 总线状态数据 id, 与总线 id 一致
  status      : 运行状态可读字符串 *
  
  last_t      : 最后采集数据时间 **
  start_t     : 首次采集时间
  inter_t     : 采集时间间隔

  data : {
    'slot_id' : {
      slot_id
      slot_desc
      dev_id 
      data_name
      data_type : 类型可读字符串
      value : 最新数据 **
      count : 计数器 **
    }
  }

  ctrl : {
    // 这个键的组合允许同一个控制槽对应不同的值
    'slot_id+value' : {
      slot_id
      slot_desc
      dev_id 
      data_name
      data_type : 类型可读字符串
      value : 初始化发送数据, 之后不变
      count : 计数器 **

      status  : 运行状态可读字符串 *
      last_t  : 最后一次发送时间 **
      start_t : 首次发送时间
      inter_t : 发送间隔时间
    }
  }
}
*/
type BusLastData struct {
  Id string `bson:"_id"`
}

const TableBusData = "bus_ldata"


/*L
授权信息
*/
type Li struct {
  AppName   string `bson:"appName"   yaml:"appName"   json:"appName"`
  Company   string `bson:"company"   yaml:"company"   json:"company"`
  Dns       string `bson:"dns"       yaml:"dns"       json:"dns"`
  Email     string `bson:"email"     yaml:"email"     json:"email"`
  BeginTime uint64 `bson:"beginTime" yaml:"beginTime" json:"beginTime"`
  EndTime   uint64 `bson:"endTime"   yaml:"endTime"   json:"endTime"`
  Z         string `bson:"z"         yaml:"z"         json:"z"`
  Api       ApiArr `bson:"api"       yaml:"api"       json:"api"`
  Signature string `bson:"signature" yaml:"signature" json:"signature"`
}

//
// Init -> ComputeZ -> Verification
//
func (l *Li) Init(yamlstr string) (error) {
  return yaml.Unmarshal([]byte(yamlstr), l)
}

func (l *Li) String() (string, error) {
  buf, err := yaml.Marshal(l)
  if err != nil {
    return "", err
  }
  return string(buf), nil
}

func (l *Li) Message() ([]byte) {
  s := l.AppName + 
       l.Company + 
       l.Dns + 
       l.Email +
       strconv.FormatUint(l.BeginTime, 10) +
       strconv.FormatUint(l.EndTime, 10) +
       Singleline(l.Z) +
       l.GetApi()
  return []byte(s)
}

func (l *Li) ComputeZ() {
  c := l.AppName + l.Company
  if c == "" {
    return
  }
  a := pick_ref_count_by_user(c)
  b := base64.StdEncoding.EncodeToString(a)
  l.Z = Multiline(b)
}

func (l *Li) GetApi() string {
  sort.Sort(&l.Api)
  return "["+ strings.Join(l.Api, ", ") +"]"
}

func (l *Li) Verification() error {
  if l.AppName != GAppName {
    return errors.New(_cpu_more_1)
  }
  p, _ := pem.Decode(pick_session_info())
  if p == nil {
    return errors.New(_cpu_more_2)
  }
  pubk, err := x509.ParsePKIXPublicKey(p.Bytes)
  if err != nil {
    return err
  }
  signed, err := base64.StdEncoding.DecodeString(l.Signature)
  if err != nil {
    return err
  }
  hash := sha1.New()
  hash.Write(l.Message())
  sum := hash.Sum([]byte{})
  return rsa.VerifyPKCS1v15(pubk.(*rsa.PublicKey), crypto.SHA1, sum, signed)
}

func (l *Li) CheckTime() error {
  now := uint64(time.Now().Unix() * 1000)
  if l.BeginTime < now && now < l.EndTime {
    return nil
  }
  return errors.New(_cpu_more_4)
}


type ApiArr []string

func (a *ApiArr) Len() int {
  return len(*a)
}

func (a *ApiArr) Less(i, j int) bool {
  return strings.Compare((*a)[i], (*a)[j]) < 0
}

func (a *ApiArr) Swap(i, j int) {
  (*a)[i], (*a)[j] = (*a)[j], (*a)[i]
}

const TableSystem = "sys_info"



/*
设备脚本

Table: dev_script {
  _id(string)       : 字典(ID)名称
  desc(string)			: 说明
  size(string)      : 脚本大小
  js                : js脚本内容
  version           : 每次修改版本+1
  cd(time)          : 创建时间
  md(time)  				: 修改时间
}
*/
type DevScript struct {
  Id      string      `bson:"_id"`
  Desc    string      `bson:"desc"`
  Size    int         `bson:"size"`
  Js      string      `bson:"js"`
  Cd      time.Time   `bson:"cd"`
  Md      time.Time   `bson:"md"`
}

const TableDevScript = "dev_script"