package service

import (
	"ic1101/brick"
	"ic1101/src/core"
)

var __i uint64 = 0
var __x uint64 = 0



func installLicenseService(b *brick.Brick) {
  ctx := &ServiceGroupContext{core.TableSystem, "软件授权"}
  __start__license(ctx)

  lserv(b, ctx, "license_get_info", license_get_info)
  lserv(b, ctx, "license_get_req",  license_get_req)
  lserv(b, ctx, "license_update",   license_update)
}


func __start__license(ctx *ServiceGroupContext) {
}


func license_get_info(h *Ht) interface{} {
  return nil
}


func license_get_req(h *Ht) interface{} {
  return nil
}


func license_update(h *Ht) interface{} {
  return nil
}