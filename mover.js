const http = require('http');
const path= require('path');
const process = require('process');
const child_process = require('child_process');
const fs = require('node:fs/promises');
const util = require('util');
const exec = util.promisify(child_process.exec);

function deliverSuccess(response) {
  response.setHeader('Content-Type', 'text/html');
  response.writeHead(200);
  response.end(`<!DOCTYPE html>
<html>
  <head>
    <title>Upload Complete</title>
  </head>
  <body>
    <p>Upload successful.</p>
  </body>
</html>
`);
}

function deliverFailure(response, error) {
  response.setHeader('Content-Type', 'text/plain');
  response.writeHead(500);
  response.end(error.toString());
}

function shellQuote(string) {
  return `'${string.replace("'", "'\\''")}'`;
}

const handleRequest = function (request, response) {
  const source = request.headers['x-source-file'];
  const destination = request.headers['x-destination-file'];
  const sed_program = '0,/^\\r$/d';
  const cleanup_command = `sed -i ${shellQuote(sed_program)} ${shellQuote(source)}`;
  exec(cleanup_command)
    .then(({stdout, stderr}) => {
      if (stdout) console.log(stdout);
      if (stderr) console.error(stderr);
      fs.mkdir(path.dirname(destination), {recursive: true})
        .then(() => fs.rename(source, destination))
            .then(() => deliverSuccess(response));
      })
    .catch(error => deliverFailure(response, error));
}

const port = 1337;
const server = http.createServer(handleRequest);
server.listen(port).on('listening', () => {
  console.log(`Listening on port ${port}.`);
});

process.on('SIGTERM', function () {
  server.close(function () { process.exit(0); });
});
