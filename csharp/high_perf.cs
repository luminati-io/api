


using System;
using System.Threading;
using System.Net;
using System.Collections.Generic;

class Client : WebClient
{
    public static string username = "lum-customer-CUSTOMER-zone-YOURZONE";
    public static string password = "YOURPASS";
    public static int port = 22225;
    public static int max_failures = 3;
    public string session_id;
    public string login;
    public string country;
    public int fail_count;
    public int n_req_for_exit_node;
    public Random rng;
    public HashSet<ServicePoint> service_points;

    public Client(string country = null)
    {
        this.country = country;
        rng = new Random();
        service_points = new HashSet<ServicePoint>();
        switch_session_id();
    }

    public void switch_session_id()
    {
        clean_connection_pool();
        session_id = rng.Next().ToString();
        n_req_for_exit_node = 0;
        update_super_proxy();
    }

    public void update_super_proxy()
    {
        Proxy = new WebProxy("session-"+session_id+".zproxy.luminati.io", port);
        login = username+(country != null ? "-country-"+country : "")
            +"-session-"+session_id;
        Proxy.Credentials = new NetworkCredential(login, password);
    }

    public void clean_connection_pool()
    {
        foreach (ServicePoint sp in service_points)
            sp.CloseConnectionGroup(login);
        service_points.Clear();
    }

    public bool have_good_super_proxy()
    {
        return fail_count < max_failures;
    }

    public void handle_response(WebException e = null) {
        if (e != null && should_switch_exit_node((HttpWebResponse)e.Response))
        {
            switch_session_id();
            fail_count++;
            return;
        }
        // success or other client/website error like 404...
        n_req_for_exit_node++;
        fail_count = 0;
    }

    public bool should_switch_exit_node(HttpWebResponse response)
    {
        return response == null ||
            status_code_requires_exit_node_switch((int)response.StatusCode);
    }

    public bool status_code_requires_exit_node_switch(int code)
    {
        return code == 403 || code == 429 || code==502 || code == 503;
    }

    protected override WebRequest GetWebRequest(Uri address)
    {
        var request = base.GetWebRequest(address) as HttpWebRequest;
        request.ConnectionGroupName = login;
        request.PreAuthenticate = true;
        return request;
    }

    protected override WebResponse GetWebResponse(WebRequest request)
    {
        var response = base.GetWebResponse(request);
        ServicePoint sp = ((HttpWebRequest)request).ServicePoint;
        service_points.Add(sp);
        return response;
    }
}

class Example
{
    public static int n_parallel_exit_nodes = 100;
    public static int n_total_req = 1000;
    public static int switch_ip_every_n_req = 20;
    public static int at_req = 0;

    static void Main()
    {
        ServicePointManager.DefaultConnectionLimit = int.MaxValue;
        for (var i = 0; i < n_parallel_exit_nodes; i++)
        {
            var t = new Thread(new ThreadStart(Run));
            t.Name = ""+i;
            t.Start();
        }
    }

    static void Run()
    {
        var client = new Client();
        while (Interlocked.Increment(ref at_req) <= n_total_req)
        {
            if (!client.have_good_super_proxy())
                client.switch_session_id();
            if (client.n_req_for_exit_node == switch_ip_every_n_req)
                client.switch_session_id();
            try {
                Console.WriteLine(client.DownloadString("http://lumtest.com/myip.json"));
                client.handle_response();
            } catch (WebException e) {
                Console.WriteLine(e.Message);
                client.handle_response(e);
            }
        }
        client.clean_connection_pool();
        client.Dispose();
    }
}

