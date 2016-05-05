#!/usr/bin/perl
use LWP::UserAgent;
my $url  = 'http://lumtest.com/myip.json';
my $user = 'lum-customer-CUSTOMER-zone-YOURZONE';
my $pass = 'YOURPASS';
my $port = 22225;
my $session = rand();
my $agent = LWP::UserAgent->new();
$agent->proxy(['http', 'https'], "http://$user-session-$session:$pass\@zproxy.luminati.io:$port");
print $agent->get($url)->content();
