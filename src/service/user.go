/**
 *  Copyright 2023 Jing Yanming
 * 
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */
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
	ctx := &ServiceGroupContext{core.TableLoginUser, "用户"}
	dserv(b, ctx, "login", 				login)
	dserv(b, ctx, "logout",  			logout)
	dserv(b, ctx, "salt",  				getsalt)
	dserv(b, ctx, "whoaim",  			whoaim)
	
	aserv(b, ctx, "reguser", 			reguser)
	aserv(b, ctx, "changepass", 	changepass)
	aserv(b, ctx, "user_list",   	user_list)
	aserv(b, ctx, "user_count",   user_count)
	aserv(b, ctx, "user_update", 	user_update)
	aserv(b, ctx, "user_delete", 	user_delete)

	mg.CreateIndex(core.TableLoginUser, &bson.D{
		{"_id", "text"}, {"weixin", "text"}, {"tel", "text"}, {"email", "text"}})
}


func encPass(name string, pass string) string {
	h := md5.New()
	h.Write([]byte(name))
	h.Write([]byte(pass))
	h.Write([]byte(salt))
  return hex.EncodeToString(h.Sum(nil))
}


func getsalt(h *Ht) interface{} {
	h.Json(HttpRet{0, "ok", salt})
	return nil
}


func login(h *Ht) interface{} {
	name := h.Get("username")
	if len(name) < 4 {
		log.Print("User Login fail: ", name)
		return errors.New("名字长度不足")
	}

	pass := h.Get("password")
	if len(pass) < 10 {
		log.Print("User Login fail: ", name)
		return errors.New("密钥长度不足")
	}

	var user *core.LoginUser
	filter := bson.D{{"_id", name}}

	if name == root.Name {
		user = &root
	} else {
		user = &core.LoginUser{}
		err := h.Table().FindOne(h.Ctx(), filter).Decode(user)
		if err != nil {
			log.Print(name, "登录失败", err)
			log.Print("User Login fail ", name)
			return errors.New("登录失败, 用户名或密码错误")
		}
	}

	pass = encPass(name, pass)
	if pass == user.Pass {
		installAuth(h, user)
		h.Session().Set("user", user)
		h.Table().UpdateOne(h.Ctx(), filter, 
				bson.D{{"$set", bson.D{{"logindata", time.Now()}}}})
				
		h.Json(HttpRet{0, "用户登录成功", nil})
		log.Print("User Login success: ", name)
	} else {
		h.Json(HttpRet{1, "用户登录失败", nil})
		log.Print("User Login fail: ", name)
	}
	return nil
}


func installAuth(h *Ht, user *core.LoginUser) {
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


func logout(h *Ht) interface{} {
	h.Session().Delete("user")
	h.Json(HttpRet{0, "用户登出", nil})
	return nil
}


func whoaim(h *Ht) interface{} {
	v := h.Session().Get("user")
	if v == nil {
		h.Json(HttpRet{1, "用户未登录", nil})
	} else {
		h.Json(HttpRet{0, "用户名", struct {
			Name string; Version string
		}{ v.(*core.LoginUser).Name, core.GVersion }})
	}
	return nil
}


func reguser(h *Ht) interface{} {
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

	_, err := h.Table().InsertOne(h.Ctx(), d)
	if err != nil {
		h.Json(HttpRet{1, "用户已经存在", err.Error()})
		return nil
	}
	h.Json(HttpRet{0, "用户已创建", name})
	return nil
}


func user_update(h *Ht) interface{} {
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

	return h.Crud().Update(id, bson.D{{"$set", d}})
}


func user_delete(h *Ht) interface{} {
	id := checkstring("用户名", h.Get("id"), 4, 64)
	return h.Crud().Delete(id)
}


func user_list(h *Ht) interface{} {
	return h.Crud().List(func(opt *options.FindOptions) {
		opt.SetProjection(bson.M{
			"_id":1, "role":1, "weixin":1, "tel":1, "email":1, 
			"isroot":1, "regdata":1, "logindata":1,
		})
	})
}


func user_count(h *Ht) interface{} {
	return h.Crud().PageInfo()
}


func changepass(h *Ht) interface{} {
	// name   := checkstring("用户名", h.Get("username"), 4, 64)
	name   := h.Session().Get("user").(*core.LoginUser).Name
	pass   := checkstring("密码", h.Get("password"), 8, 64)
	oldpass:= checkstring("旧密码", h.Get("oldpassword"), 8, 64)

	filter := bson.D{{"_id", name}}
	user   := core.LoginUser{}
	err    := h.Table().FindOne(h.Ctx(), filter).Decode(&user)
	if err != nil {
		return err
	}

	oldpass = encPass(name, oldpass)
	if oldpass != user.Pass {
		return errors.New("旧密码错误")
	}

	h.Table().UpdateOne(h.Ctx(), filter, 
			bson.D{{"$set", bson.D{{"pass", encPass(name, pass)}} }})

	h.Json(HttpRet{0, "密码已修改", name})
	return nil
}