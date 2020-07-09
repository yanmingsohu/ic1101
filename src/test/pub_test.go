package test

import (
	"fmt"
	"ic1101/src/core"
	"testing"
)

const _cpu_core_info = `
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA3RRpu+CFRgTajKY1g0Sz
SXH51p6uphRuPN/Jh0ZZV8vdQVB9LYT9xYLgDmm2AU72BU1f7GdNp7o6yML7meKz
TgHzRRa6cf1pqwrfYeD2yVITfNssmo4njimM/elY/K6ukTbR8lbbjCj2i7SoYwzn
ib2edNSc6yj6I4R61proS7v7kAfIuQn5/a8hdVPgg30JHWw+hDAQkyh2MTZYDdQq
C8N+tlEDKK7wFIkAiUyPqZSSkBo5mwtkvoXopE8VxUxJAePlOS8csMMCINF2uekB
diVtlH84RxIw8442LXUvXJJDiboh2pKSJyYDQZ8/UX2BOjg0kzHbq1+3wFiHd3wJ
aQIDAQAB
-----END PUBLIC KEY-----
`

const _yaml = `
appName: 智慧大数据开放平�?
company: 上海竹呗信息�?�?
dns: zhubei.com
email: zhubei.com
beginTime: 1527487634019
endTime: 1843106834000
z: |-
   2ABY8B2lr07bvL9Nw64H8++1qvMPXxoHvJy6jFxqlMOUbSe7y4
   J5tKvBEo8WSlS7nn044yB/YXhPbnNabt+99g==
api: 
- app.module.cluster.functions()
- api.ide.code.modify.functions()
- app.module.chain.peer.platform()
- app.module.fabric.functions()
- app.module.webservice.functions()
- app.module.schedule.functions()
- app.module.mongodb.functions()
- app.module.apipm.functions()
- app.module.shell.functions()
signature: |-
   RVk1KiE14bWej7yjTSHjORsVOtWbqU6nhkMhUUzthMfIaGirh/
   4s7Tf2tzvnt9VpvSuoVX/ddEVPJckH54VXuYxcDC8xFLeksmfi
   94VBfnaw4Uj5d3Zh+7GjxDNmtxv5miwaFaK5nhyx/KtsB37ffq
   hztzd2fePgMd20YXv4v7TXo88JwHdBpxAcV9i3TvWQzvifyd9x
   vSqi9ZdruFIb9oZDtVbTXqUJ5Mft199Hu5KgLw10nCtTgbz4Y9
   0Je4GkKK6sgWK9Z76Ww2I8B2U7gvkYLW9OuFG//k2lk/M9htZU
   cChDkkYBRFKrMwwENlyd4X55FDpQWuNS1dj0iLF65A==
`

func TestReadPub(t *testing.T) {
  li := core.Li{}
  li.Init(_yaml)

  fmt.Println("V:", li.Verification())
}