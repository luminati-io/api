#!/usr/bin/env node
var request = require('request-promise');
var username = 'lum-customer-CUSTOMER-zone-YOURZONE';
var password = 'YOURPASS';
var port = 22225;
var session_id = (1000000 * Math.random())|0;
var super_proxy = 'http://'+username+'-session-'+session_id+':'+password+'@zproxy.luminati.io:'+port;
var options = {
    url: 'http://lumtest.com/myip.json',
    proxy: super_proxy,
};
console.log('Performing request');
request(options)
.then(function(data){ console.log(data); }, function(err){ console.error(err); });
