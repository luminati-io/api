<?php
$username = 'lum-customer-CUSTOMER-zone-YOURZONE';
$password = 'YOURPASS';
$port = 22225;
$session = mt_rand();
$super_proxy = 'zproxy.luminati.io';
$curl = curl_init('http://lumtest.com/myip.json');
curl_setopt($curl, CURLOPT_RETURNTRANSFER, 1);
curl_setopt($curl, CURLOPT_PROXY, "http://$super_proxy:$port");
curl_setopt($curl, CURLOPT_PROXYUSERPWD, "$username-session-$session:$password");
$result = curl_exec($curl);
curl_close($curl);
if ($result)
    echo $result;
?>
