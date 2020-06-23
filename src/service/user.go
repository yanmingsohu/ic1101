package service

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"log"
	"time"

	"../../brick"
	"../core"
	"go.mongodb.org/mongo-driver/bson"
)


type User struct {
	Name string
	Auth interface{}
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
		//TODO: 加载权限
		h.Session().Set("user", User{name, nil})
		table.UpdateOne(h.Ctx(), filter, 
				bson.D{{"$set", bson.D{{"logindata", time.Now()}}}})
		h.Json(HttpRet{0, "用户登录成功", nil})
	} else {
		h.Json(HttpRet{1, "用户登录失败", nil})
	}
	return nil
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
		h.Json(HttpRet{0, "用户名", v.(User).Name})
	}
	return nil
}


func reguser(h brick.Http) error {
	now 	 := time.Now()
	pass   := checkstring("密码", h.Get("password"), 8, 64)
	name   := checkstring("用户名", h.Get("username"), 4, 64)

	d := bson.D{
		{"_id",   		name},
		{"pass",  		encPass(name, pass)},
		{"wexin", 		h.Get("wexin")},
		{"tel",   		h.Get("tel")},
		{"email", 		h.Get("email")},
		{"regdata",   now},
		{"logindata", now},
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


func changepass(h brick.Http) error {
	name   := checkstring("用户名", h.Get("username"), 4, 64)
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
			bson.D{{"$set", bson.D{{"pass", encPass(name, pass)}}}})

	h.Json(HttpRet{0, "密码已修改", name})
	return nil
}