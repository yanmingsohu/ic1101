package service

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"log"
	"time"

	"ic1101/brick"
	"ic1101/src/core"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)


func installUserService(b *brick.Brick) {
	dserv(b, "login", 			login)
	dserv(b, "logout",  		logout)
	dserv(b, "salt",  			getsalt)
	dserv(b, "whoaim",  		whoaim)
	
	aserv(b, "reguser", 		reguser)
	aserv(b, "changepass", 	changepass)
	aserv(b, "user_list",   user_list)
	aserv(b, "user_update", user_update)

	mg.CreateIndex("dict", &bson.D{
		{"_id", "text"}, {"weixin", "text"}, {"tel", "text"}, {"email", "text"}})
}


func encPass(name string, pass string) string {
	h := md5.New()
	h.Write([]byte(name))
	h.Write([]byte(pass))
	h.Write([]byte(salt))
  return hex.EncodeToString(h.Sum(nil))
}


func getsalt(h brick.Http) error {
	h.Json(HttpRet{0, "ok", salt})
	return nil
}


func login(h brick.Http) error {
	name := h.Get("username")
	if len(name) < 4 {
		return errors.New("名字长度不足")
	}

	pass := h.Get("password")
	if len(pass) < 10 {
		return errors.New("密钥长度不足")
	}

	var user *core.LoginUser
	filter := bson.D{{"_id", name}}
	table := mg.Collection("login_user")

	if name == root.Name {
		user = &root
	} else {
		user = &core.LoginUser{}
		err := table.FindOne(h.Ctx(), filter).Decode(user)
		if err != nil {
			log.Print(name, "登录失败", err)
			return errors.New("登录失败, 用户名或密码错误")
		}
	}

	pass = encPass(name, pass)
	if pass == user.Pass {
		installAuth(h, user)
		h.Session().Set("user", user)
		table.UpdateOne(h.Ctx(), filter, 
				bson.D{{"$set", bson.D{{"logindata", time.Now()}}}})
		h.Json(HttpRet{0, "用户登录成功", nil})
	} else {
		h.Json(HttpRet{1, "用户登录失败", nil})
	}
	return nil
}


func installAuth(h brick.Http, user *core.LoginUser) {
	if user.Role == "" {
		return
	}
	ruels, err := getRuels(h, user.Role)

	if err != nil {
		log.Print("Install User ", user.Name, " Auth fail ", err)
		return
	}

	for _, v := range ruels {
		user.Auths[v] = true
	}
}


func logout(h brick.Http) error {
	h.Session().Delete("user")
	h.Json(HttpRet{0, "用户登出", nil})
	return nil
}


func whoaim(h brick.Http) error {
	v := h.Session().Get("user")
	if v == nil {
		h.Json(HttpRet{1, "用户未登录", nil})
	} else {
		h.Json(HttpRet{0, "用户名", v.(*core.LoginUser).Name})
	}
	return nil
}


func reguser(h brick.Http) error {
	now 	 := time.Now()
	name   := checkstring("用户名", h.Get("username"), 4, 64)
	pass   := checkstring("密码", h.Get("password"), 8, 64)
	isroot := checkbool("超级用户", h.Get("rootuser"))

	if isroot {
		if !h.Session().Get("user").(*core.LoginUser).IsRoot {
			return errors.New("只有超级用户可以创建另一个超级用户")
		}
	}

	d := bson.D{
		{"_id",   		name},
		{"pass",  		encPass(name, pass)},
		{"weixin", 		h.Get("weixin")},
		{"tel",   		h.Get("tel")},
		{"email", 		h.Get("email")},
		{"regdata",   now},
		{"logindata", now},
		{"isroot",    isroot},
		{"role",      h.Get("role")},
	}

	table := mg.Collection("login_user")
	_, err := table.InsertOne(h.Ctx(), d)
	if err != nil {
		h.Json(HttpRet{1, "用户已经存在", err.Error()})
		return nil
	}
	h.Json(HttpRet{0, "用户已创建", name})
	return nil
}


func user_update(h brick.Http) error {
	id := checkstring("用户名", h.Get("username"), 4, 64)
	user := h.Session().Get("user").(*core.LoginUser)

	if !user.IsRoot {
		roles, err := getRuels(h, user.Role)
		if err != nil {
			return err
		}

		for _, id := range roles {
			if !user.Auths[id] {
				return errors.New("当前用户无权赋予角色, 缺少权限: "+ id)
			}
		}
	}

	d := bson.D{
		{"weixin", 		h.Get("weixin")},
		{"tel",   		h.Get("tel")},
		{"email", 		h.Get("email")},
		{"role",      h.Get("role")},
	}

	c := Crud{h, "login_user", "用户"}
	return c.Update(id, bson.D{{"$set", d}})
}


func user_list(h brick.Http) error {
	c := Crud{h, "login_user", "用户"}
	return c.List(func(opt *options.FindOptions) {
		opt.SetProjection(bson.M{
			"_id":1, "role":1, "weixin":1, "tel":1, "email":1, 
			"isroot":1, "regdata":1, "logindata":1,
		})
	})
}


func changepass(h brick.Http) error {
	// name   := checkstring("用户名", h.Get("username"), 4, 64)
	name   := h.Session().Get("user").(*core.LoginUser).Name
	pass   := checkstring("密码", h.Get("password"), 8, 64)
	oldpass:= checkstring("旧密码", h.Get("oldpassword"), 8, 64)

	filter := bson.D{{"_id", name}}
	table  := mg.Collection("login_user")
	user   := core.LoginUser{}
	err    := table.FindOne(h.Ctx(), filter).Decode(&user)
	if err != nil {
		return err
	}

	oldpass = encPass(name, oldpass)
	if oldpass != user.Pass {
		return errors.New("旧密码错误")
	}

	table.UpdateOne(h.Ctx(), filter, 
			bson.D{{"$set", bson.D{{"pass", encPass(name, pass)}} }})

	h.Json(HttpRet{0, "密码已修改", name})
	return nil
}