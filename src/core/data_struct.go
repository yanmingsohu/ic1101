package core

import "go.mongodb.org/mongo-driver/bson"

const TimeFormatString = "2020-06-25T08:32:44.676+00:00"

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
  Name string 			`yaml:"username" bson:"_id"`
  Pass string 			`yaml:"password" bson:"pass"`
  Role string			 	`bson:"role"`
  IsRoot bool      	`bson:"isroot"`
  Auths map[string]bool
}

const TableLoginUser = "login_user"


/*
Table: dict {
  _id(string)       : 字典(ID)名称
  desc(string)			: 说明
  content           : 字典内容
  cd                : 创建时间
  md  							: 修改时间
}
dict: content {
  key : value
}
*/
type Dict struct {
  Id 			string  					`bson:"_id"`
  Desc 		string 						`bson:"desc"`
  Cd      string            `bson:"cd"`
  Md      string            `bson:"md"`
  Content map[string]string `bson:"content"`
}

const TableDict = "dict"


/*
Table: role {
  _id(string)     : 角色名
  desc(string)    : 说明
  cd              : 创建时间
  md              : 修改时间
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
  cd             : 创建时间
  md             : 修改时间
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
  Id        string  `bson:"_id"`
  Desc      string  `bson:"desc"`
  ChangeId  int     `bson:"changeid"`
  Cd        string  `bson:"cd"`
  Md        string  `bson:"md"`
  Script    string  `bson:"script"`

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
  md        : 修改时间
  cd        : 创建时间
  
  dd        : 最后数据时间
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
    v : (数据 map) {
      Y : n年数据, (数字类型)
      ...
    }
  }

  当年所有月数据: {
    _id : month$Y
    v : {
      1 : 1月数据 (月份是数字类型)
      ...
      12 : 12月数据
    }
  }

  日数据: {
    _id : day$Y-M
    v : {
      1 : x月1日数据
      ...
      31 : 31日数据
    }
  }

  小时数据: {
    _id : hour$Y-M-D
    v : {
      0 : 0点数据
      ...
      23 : 23点数据
    }
  }

  分钟数据:  {
    _id : minute$Y-M-D_h
    v : {
      0 : 0分数据
      ...
      59: 59分数据
    }
  }

  秒数据:  {
    _id : second$Y-M-D_h:m
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