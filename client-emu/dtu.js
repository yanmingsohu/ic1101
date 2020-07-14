const net = require("net");


runSlave(1);
runSlave(9);


function runSlave(id) {
  console.log("模拟从机 ", id);
  const dirty = Buffer.from([0xfe, 0, id, id, 0xfe])

  const client = net.createConnection({ port: 502 }, () => {
    // 'connect' 监听器
    client.write(dirty);
    console.log("Header", id);
  });

  client.on('data', (data) => {
    console.log("GET ", id, '<', hex(data));
    let sess = data.readUInt16BE(0);
    let prot = data.readUInt16BE(2);
    let len  = data.readUInt16BE(4);
    let sid  = data.readUInt8(6);
    let func = data.readUInt8(7);
    let addr = data.readUInt16BE(8);
    let cout = data.readUInt16BE(10);
    // console.log(sess, prot, len, sid, func, addr, cout);
    
    let w = Buffer.alloc(9 + (cout<<1));
    w.writeUInt16BE(sess, 0);
    w.writeUInt16BE(prot, 2);
    w.writeUInt16BE((cout<<1) + 3, 4)
    w.writeUInt8(sid, 6);
    w.writeUInt8(func, 7);
    w.writeUInt8(cout<<1, 8);
    for (let i=0; i<cout; ++i) {
      w.writeUInt16BE(Math.random() * 100, 9+(i<<1));
    }

    console.log("SEND", id, '>', hex(w));
    client.write(w);
  });

  client.on('end', () => {
    console.log('已从服务器断开');
  });
}


function hex(buf) {
  let str = [];
  for (let i=0; i<buf.length; ++i) {
    let a = buf[i];
    if (a < 0xf) {
      str.push('0'+ a.toString(16));
    } else {
      str.push(a.toString(16));
    }
  }
  return str.join(" ");
}