const http = require('http');

const server = http.createServer((req, res) => {
  console.log("GET REQ:", req.url)

  req.on("data", function(d) {
    console.log("BODY:", d, d.toString('utf8'));
  });

  req.on("end", function() {
    res.writeHead(200, { 'Content-Type': 'text/plain' });
    res.end('---bye---');
  });
});

const port = 90;
server.listen(port, '127.0.0.1');
console.log("Http server on", port);