#!/usr/bin/env node
var request = require('request-promise');
var promise = require('bluebird'); // promises lib used by request-promise
var lookup = promise.promisify(require('dns').lookup);
var http = require('http');
var username = 'lum-customer-CUSTOMER-zone-YOURZONE';
var password = 'YOURPASS';
var port = 22225;
var at_req = 0;
var n_total_req = 1000;
var n_parallel_exit_nodes = 100;
var switch_ip_every_n_req = 50;
var max_failures = 3;
var req_timeout = 60*1000;

function main(){
    http.Agent.defaultMaxSockets = Infinity;
    for (var i=0; i < n_parallel_exit_nodes; i++)
        new Session(i).start();
}

function Session(id){
    this.id = id;
    this.n_req_for_exit_node = 0;
    this.fail_count = 0;
    this.switch_session_id();
}

var proto = Session.prototype;

proto.start = proto.next = function(){
    if (at_req >= n_total_req)
        return this.cleanup(); // all done
    at_req++;
    var _this = this;
    promise.try(function(){
        if (!_this.have_good_super_proxy())
            return _this.switch_super_proxy();
    }).then(function(){
        if (_this.n_req_for_exit_node==switch_ip_every_n_req)
            _this.switch_session_id();
        var options = {
            url: 'http://lumtest.com/myip.json',
            timeout: req_timeout,
            pool: _this.pool,
            forever: true,
            proxy: _this.super_proxy_url,
        };
        return request(options);
    }).then(function success(res){
        console.log(res);
        _this.fail_count = 0;
        _this.n_req_for_exit_node++;
    }, function error(err){
        if (err.statusCode
            && !status_code_requires_exit_node_switch(err.statusCode))
        {
            // this could be 404 or other website error
            _this.n_req_for_exit_node++;
            return;
        }
        _this.switch_session_id();
        _this.fail_count++;
    }).finally(function(){
        _this.next();
    });
};

proto.have_good_super_proxy = function(){
    return this.super_proxy_host && this.fail_count < max_failures;
};

proto.update_super_proxy_url = function(){
    this.super_proxy_url = 'http://'+username+
        '-session-'+
        this.session_id+':'+password+'@'+this.super_proxy_host+':'+port;
};

proto.switch_session_id = function(){
    connection_pool_cleanup(this.pool);
    this.pool = {};
    this.session_id = (1000000 * Math.random())|0;
    this.n_req_for_exit_node = 0;
    this.update_super_proxy_url();
};

proto.switch_super_proxy = function(){
    var _this = this;
    this.switch_session_id();
    return promise.try(function(){
        return lookup('session-'+_this.session_id+
            '.'+
            'zproxy.luminati.io');
    }).then(function success(res){
        _this.super_proxy_host = res;
        _this.update_super_proxy_url();
    });
};

proto.cleanup = function(){
    connection_pool_cleanup(this.pool);
};

function connection_pool_cleanup(pool){
    if (!pool)
        return;
    Object.keys(pool).forEach(function(key){
        var sockets = pool[key].sockets;
        Object.keys(sockets).forEach(function(name){
            sockets[name].forEach(function(s){
                s.destroy();
            });
        });
    });
}

function status_code_requires_exit_node_switch(status_code){
    return [403, 429, 502, 503].indexOf(status_code)>=0;
}

main();

