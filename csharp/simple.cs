
using System;
using System.Net;

class Client : WebClient
{
    public static string username = "lum-customer-CUSTOMER-zone-YOURZONE";
    public static string password = "YOURPASS";
    public static int port = 22225;
    public string session_id = new Random().Next().ToString();

    public Client(string country = null)
    {
        this.Proxy = new WebProxy("zproxy.luminati.io", port);
        var login = username+(country != null ? "-country-"+country : "")
            +"-session-"+session_id;
        this.Proxy.Credentials = new NetworkCredential(login, password);
    }

    protected override WebRequest GetWebRequest(Uri address)
    {
        var request = base.GetWebRequest(address) as HttpWebRequest;
        request.ConnectionGroupName = session_id;
        return request;
    }
}

class Example
{
    static void Main()
    {
        Console.WriteLine("Performing request(s)");
        var client = new Client();
        // Put full scraping sequence below:
        Console.WriteLine(client.DownloadString("http://lumtest.com/myip.json"));
        // client.DownloadString(...second request...);
    }
}



