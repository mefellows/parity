// app.js
var express = require('express');
var path = require('path');
var http = require('http');
var app = module.exports.app = exports.app = express();
var publicDir = path.join(__dirname, '')

//you won't need 'connect-livereload' if you have livereload plugin for your browser
app.set('port', process.env.PORT || 3000)
app.use(require('connect-livereload')());
app.get('/', function(req, res) {
  res.sendFile(path.join(publicDir, 'index.html'))
})

var server = http.createServer(app)

server.listen(app.get('port'), function(){
  console.log("Web server listening on port " + app.get('port'));
});
