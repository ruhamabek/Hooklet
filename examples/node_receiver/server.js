// Zero-dependency Node.js Webhook Receiver (runs on any Node.js version)
const http = require('http');

const server = http.createServer(async (req, res) => {
  let body = '';
  for await (const chunk of req) {
    body += chunk;
  }

  console.log(`\n [Node.js :8000] Received ${req.method} on ${req.url}`);
  console.log(` X-Hooklet-Replayed: ${req.headers['x-hooklet-replayed'] || 'false'}`);
  console.log(` Signature Header:   ${req.headers['stripe-signature'] || req.headers['x-hub-signature-256'] || 'none'}`);
  console.log(` Payload Bytes:      ${body.length}`);
  if (body) {
    try {
      console.log('JSON Body:\n', JSON.stringify(JSON.parse(body), null, 2));
    } catch {
      console.log(' Raw Body:', body);
    }
  }

  res.writeHead(200, { 'Content-Type': 'application/json' });
  res.end(JSON.stringify({
    status: 'success',
    replayed_by_hooklet: req.headers['x-hooklet-replayed'] === 'true',
    received_at: new Date().toISOString()
  }));
});

const PORT = 8000;
server.listen(PORT, () => {
  console.log(`\nLocal Webhook Receiver running on http://localhost:${PORT}`);
  console.log(`   Ready to receive replays from Hooklet!\n`);
});
