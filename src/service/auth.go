package service

import (
	"errors"
	"ic1101/brick"
	"ic1101/src/core"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)


func installAuthService(b *brick.Brick) {
  ctx := &ServiceGroupContext{"role", "角色"}
  aserv(b, ctx, "auth_list",       auth_list)
  aserv(b, ctx, "role_count",      role_count)
  aserv(b, ctx, "role_list",       role_list)
  aserv(b, ctx, "role_create",     role_create)
  aserv(b, ctx, "role_delete",     role_delete)
  aserv(b, ctx, "role_read_rule",  role_read_rule)
  aserv(b, ctx, "role_update",     role_update)

  mg.CreateIndex("role", &bson.D{{"_id", "text"}, {"desc", "text"}})
}


func auth_list(h *Ht) interface{} {
  h.Json(HttpRet{0, "auth list", auth_arr})
  return nil
}


func role_count(h *Ht) interface{} {
  c := Crud{h, "role", "角色"}
  return c.PageInfo()
}


func role_list(h *Ht) interface{} {
  c := Crud{h, "role", "角色"}
  return c.List(func(opt *options.FindOptions) {
    opt.SetProjection(bson.M{"desc":1, "cd":1, "md":1})
  })
}


func role_create(h *Ht) interface{} {
  d := bson.D{
    {"_id",  checkstring("角色ID", h.Get("id"), 2, 20)},
    {"desc", checkstring("角色说明", h.Get("desc"), 0, 999)},
    {"cd", 	 time.Now()},
    {"md",   ""},
  }
  c := Crud{h, "role", "角色"}
  return c.Create(&d)
}


func role_delete(h *Ht) interface{} {
  id := checkstring("角色ID", h.Get("id"), 2, 20)
  c := Crud{h, "role", "角色"}
  return c.Delete(id)
}


func role_update(h *Ht) interface{} {
  user := h.Session().Get("user").(*core.LoginUser)
  id := checkstring("角色ID", h.Get("id"), 2, 20)
  c := Crud{h, "role", "角色"}

  up := bson.M{
    "desc" : h.Get("desc"),
    "md"   : time.Now(),
  }

  rules := h.Gets("r")
  if len(rules) > 0 {
    for _, id := range rules {
      if !user.Auths[id] {
        return errors.New("当前用户不能赋予权限: "+ id)
      }
    }
    up["rules"] = rules
  }

  return c.Update(id, bson.M{"$set": up})
}


func role_read_rule(h *Ht) interface{} {
  id := checkstring("角色ID", h.Get("id"), 2, 20)
  c := Crud{h, "role", "角色"}
  return c.Read(id, "rules")
}


func getRuels(h *Ht, ruleId string) ([]string, error) {
	filter := bson.D{{"_id", ruleId}}
	opt    := options.FindOne()
	ret    := bson.M{}
	opt.SetProjection(bson.M{"rules": 1})
	err    := mg.Collection("auth").FindOne(h.Ctx(), filter, opt).Decode(&ret)

	if err != nil {
		return nil, err
	}
	return ret["rules"].([]string), nil
}