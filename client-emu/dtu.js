const net = require("net");

const slaveID = 1;

const client = net.createConnection({ port: 502 }, () => {
  // 'connect' 监听器
  client.write(Buffer.from([0xfe, 0, slaveID, slaveID, 0xfe]));
  console.log("write");
});

client.on('data', (data) => {
  console.log(data.toString());
  client.end();
});

client.on('end', () => {
  console.log('已从服务器断开');
});