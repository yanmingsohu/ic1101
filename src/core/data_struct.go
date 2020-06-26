package core

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


/*
Table: dev-proto {
  _id(string)    : 原型id
  desc(string)   : 说明
  cd             : 创建时间
  md             : 修改时间
  dd             : 最后数据时间
  dc(int64)      : 数据量

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
      script   : '脚本名, 用于处理虚拟数据'
    }
  ]

  ctrls : 控制列表 [
    { name    : '名称'
      desc    : '说明'
      type    : DevDataType '发送的控制数据'
      script  : '脚本名, 用于发送虚拟控制'
    }
  ]
}
*/
type DevProto struct {
  Id string `bson:"_id"`
}

type DevAttrType int
type DevDataType int

const (
  DDT_int       DevDataType = 1 // 整数类型
  DDT_float     DevDataType = 2 // 浮点类型
  DDT_virtual   DevDataType = 3 // 虚拟数据
  DDT_sw        DevDataType = 4 // 开关类型

  DAT_string    DevAttrType = 100 // 字符串
  DAT_number    DevAttrType = 101 // 数字
  DAT_dict      DevAttrType = 102 // 字典
)