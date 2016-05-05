#!/bin/bash
USERNAME=lum-customer-CUSTOMER-zone-YOURZONE
PASSWORD=YOURPASS
PORT=22225
echo "Choosing fastest Super Proxy"
SESSION=$RANDOM
echo "Performing request"
curl --proxy zproxy.luminati.io:$PORT \
     --proxy-user $USERNAME-session-$SESSION:$PASSWORD \
     "http://lumtest.com/myip.json"
