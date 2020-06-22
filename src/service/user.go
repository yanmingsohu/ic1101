package service

import (
	"crypto/md5"
	"encoding/hex"
	"errors"

	"../../brick"
)


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

	if name == root.Name {
		pass := encPass(name, pass)
		if pass == root.Pass {
			h.Session().Set("user", User{name, nil})
			h.Json(HttpRet{0, "ROOT 用户登录成功", nil})
		} else {
			h.Json(HttpRet{1, "ROOT 用户登录失败", nil})
		}
		return nil
	}

	return errors.New("not implement "+ name + pass)
}