<?php
$username = 'lum-customer-CUSTOMER-zone-YOURZONE';
$password = 'YOURPASS';
$port = 22225;

class Mcurl
{
    private $mh;
    private $curl_handlers;
    public function __construct(){
        $this->mh = curl_multi_init();
        $this->curl_handlers = array();
    }
    public function __destruct(){
        curl_multi_close($this->mh);
    }
    public function async_get($curl_options, $handler){
        $curl = curl_init();
        curl_setopt_array($curl, $curl_options);
        $this->async_exec($curl, $handler);
    }
    private function async_exec($curl, $handler){
        $this->curl_handlers[(string)$curl] = $handler;
        curl_multi_add_handle($this->mh, $curl);
    }
    public function run(){
        $active = 0;
        $mrc = 0;
        do {
            $mrc = curl_multi_exec($this->mh, $active);
        } while ($mrc == CURLM_CALL_MULTI_PERFORM);
        while ($active && $mrc == CURLM_OK) {
            $mrc = curl_multi_select($this->mh);
            do {
                $mrc = curl_multi_exec($this->mh, $active);
            } while ($mrc == CURLM_CALL_MULTI_PERFORM);
            while ($info = curl_multi_info_read($this->mh)) {
                $curl = $info['handle'];
                $key = (string)$curl;
                $handler = $this->curl_handlers[$key];
                unset($this->curl_handlers[$key]);
                $info = curl_getinfo($curl);
                $http_code = $info['http_code'];
                $content = curl_multi_getcontent($curl);
                //echo "http_code:$http_code curl:$curl content: $content\n";
                $handler($http_code, $content);
                curl_multi_remove_handle($this->mh, $curl);
                curl_close($curl);
            }
        }
    }
};

class Client
{
    public $super_proxy;
    public $session_id;
    public $fail_count;
    public $n_req_for_exit_node;
    public $mcurl;
    private $proxy, $auth;

    public function __construct($mcurl)
    {
        $this->session_id = "";
        $this->fail_count = 0;
        $this->n_req_for_exit_node = 0;
        $this->mcurl = $mcurl;
        $this->switch_session_id();
    }
    private function switch_session_id(){
        $this->session_id = mt_rand();
        #echo "switched session ID to: ".$this->session_id."\n\n";
        $this->n_req_for_exit_node = 0;
        $this->update_super_proxy();
    }
    private function update_super_proxy(){
        global $port, $username, $password;
        $this->fail_count = 0;
        $super_proxy = "session-".$this->session_id.".zproxy.luminati.io";
        $this->proxy = "http://".$super_proxy.":$port";
        $this->auth = "$username-session-".$this->session_id.":$password";
    }
    private function have_good_super_proxy(){
        global $max_failures;
        return $this->fail_count < $max_failures;
    }
    private function make_request(){
        $curl_options = array(
            CURLOPT_URL => 'http://lumtest.com/myip.json',
            CURLOPT_RETURNTRANSFER => 1,
	        CURLOPT_PROXY => $this->proxy,
	        CURLOPT_PROXYUSERPWD => $this->auth,
        );
        $client = $this;
        $handler = function($http_code, $content) use ($client){
            $client->handle_response($http_code, $content); };
        $this->mcurl->async_get($curl_options, $handler);
    }
    private function handle_response($http_code, $content){
        if ($this->should_switch_exit_node($http_code, $content)){
           $this->switch_session_id();
           $this->fail_count++;
           $this->run_next();
        }else{
            // success or other client/website error like 404...
            echo "$content\n";
            $this->n_req_for_exit_node++;
            $this->fail_count = 0;
            $this->run();
        }
    }
    private function should_switch_exit_node($http_code, $content){
        return $content=="" ||
            $this->status_code_requires_exit_node_switch($http_code);
    }
    private function status_code_requires_exit_node_switch($code){
        if (!$code) // curl_multi timed out or failed
            return true;
        return $code==403 || $code==429 || $code==502 || $code==503;
    }
    private function run_next(){
        global $switch_ip_every_n_req;
        if (!$this->have_good_super_proxy()){
            $this->switch_session_id();
            return;
        }
        if ($this->n_req_for_exit_node == $switch_ip_every_n_req)
            $this->switch_session_id();
        $this->make_request();
    }
    public function run(){
        global $n_total_req, $at_req;
        if ($at_req++ < $n_total_req){
            $this->run_next();
        }
    }
};

$mcurl = new Mcurl();

$max_failures = 3;
$n_parallel_exit_nodes = 10;
$n_total_req = 500;
$switch_ip_every_n_req = 20;
$at_req = 0;

for ($i=0; $i<$n_parallel_exit_nodes; ++$i){
    $client = new Client($mcurl);
    $client->run();
}
$mcurl->run();
?>

