package core

/*
Table: user {
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