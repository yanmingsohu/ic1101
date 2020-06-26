package service

import (
	"ic1101/brick"

	"go.mongodb.org/mongo-driver/bson"
)


func installDevProtoService(b *brick.Brick) {
  mg.CreateIndex("dev-proto", &bson.D{{"_id", "text"}, {"desc", "text"}})
  ctx := &ServiceGroupContext{"dev-proto", "设备原型"}

  aserv(b, ctx, "dev_proto_create", dev_proto_create)
}


func dev_proto_create(h *Ht) interface{} {
  return nil
}