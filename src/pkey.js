const z = require('zlib')

const v = {
  _cpu_core_info : `
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA3RRpu+CFRgTajKY1g0Sz
SXH51p6uphRuPN/Jh0ZZV8vdQVB9LYT9xYLgDmm2AU72BU1f7GdNp7o6yML7meKz
TgHzRRa6cf1pqwrfYeD2yVITfNssmo4njimM/elY/K6ukTbR8lbbjCj2i7SoYwzn
ib2edNSc6yj6I4R61proS7v7kAfIuQn5/a8hdVPgg30JHWw+hDAQkyh2MTZYDdQq
C8N+tlEDKK7wFIkAiUyPqZSSkBo5mwtkvoXopE8VxUxJAePlOS8csMMCINF2uekB
diVtlH84RxIw8442LXUvXJJDiboh2pKSJyYDQZ8/UX2BOjg0kzHbq1+3wFiHd3wJ
aQIDAQAB
-----END PUBLIC KEY-----
`,

  _cpu_more_1 : "无效的应用名称",
  _cpu_more_2 : "公钥损坏",
  C_cpu_mre_3 : "应用需要授权",
  _cpu_more_4 : "授权过期",
}



const buf = ['func init() {\n var b []byte'];

for (let n in v) {
  buf.push('\n b = UnZip([]byte{');
  let chunk = z.gzipSync(v[n]);

  for (var i=0; i<chunk.length; ++i) {
    var b = chunk[i];
    buf.push(b.toString());
    buf.push(',');
    if (i%20 == 0) buf.push('\n  ');
  }
  buf.push("\n })");
  buf.push("\n ", n, " = string(b)");
}

buf.push("\n}");

console.log(buf.join(''));