package example;

import java.io.*;
import java.util.Random;
import java.util.concurrent.*;
import java.util.concurrent.atomic.AtomicInteger;

import org.apache.http.HttpHost;
import org.apache.http.HttpResponse;
import org.apache.http.auth.*;
import org.apache.http.client.CredentialsProvider;
import org.apache.http.client.config.RequestConfig;
import org.apache.http.client.fluent.Request;
import org.apache.http.client.methods.*;
import org.apache.http.impl.client.*;
import org.apache.http.impl.conn.BasicHttpClientConnectionManager;
import org.apache.http.impl.conn.PoolingHttpClientConnectionManager;
import org.apache.http.util.EntityUtils;

class Client {
    public static final String username = "lum-customer-CUSTOMER-zone-YOURZONE";
    public static final String password = "YOURPASS";
    public static final int port = 22225;
    public static final int max_failures = 3;
    public static final int req_timeout = 60*1000;
    public String session_id;
    public HttpHost super_proxy;
    public CloseableHttpClient client;
    public String country;
    public int fail_count;
    public int n_req_for_exit_node;
    public Random rng;

    public Client(String country) {
        this.country = country;
        rng = new Random();
        switch_session_id();
    }

    public void switch_session_id() {
        session_id = Integer.toString(rng.nextInt(Integer.MAX_VALUE));
        n_req_for_exit_node = 0;
        super_proxy = new HttpHost("session-"+session_id+".zproxy.luminati.io", port);
        update_client();
    }

    public void update_client() {
        close();
        String login = username+(country!=null ? "-country-"+country : "")
                +"-session-" + session_id;
        CredentialsProvider cred_provider = new BasicCredentialsProvider();
        cred_provider.setCredentials(new AuthScope(super_proxy),
                new UsernamePasswordCredentials(login, password));
        RequestConfig config = RequestConfig.custom()
                .setConnectTimeout(req_timeout)
                .setConnectionRequestTimeout(req_timeout)
                .build();
        PoolingHttpClientConnectionManager conn_mgr =
                new PoolingHttpClientConnectionManager();
        conn_mgr.setDefaultMaxPerRoute(Integer.MAX_VALUE);
        conn_mgr.setMaxTotal(Integer.MAX_VALUE);
        client = HttpClients.custom()
                .setConnectionManager(conn_mgr)
                .setProxy(super_proxy)
                .setDefaultCredentialsProvider(cred_provider)
                .setDefaultRequestConfig(config)
                .build();
    }

    public CloseableHttpResponse request(String url) throws IOException {
        try {
            HttpGet request = new HttpGet(url);
            CloseableHttpResponse response = client.execute(request);
            handle_response(response);
            return response;
        } catch (IOException e) {
            handle_response(null);
            throw e;
        }
    }

    public void handle_response(HttpResponse response) {
        if (response != null && !status_code_requires_exit_node_switch(
                response.getStatusLine().getStatusCode())) {
            // success or other client/website error like 404...
            n_req_for_exit_node++;
            fail_count = 0;
            return;
        }
        switch_session_id();
        fail_count++;
    }

    public boolean status_code_requires_exit_node_switch(int code) {
        return code == 403 || code == 429 || code==502 || code == 503;
    }

    public boolean have_good_super_proxy() {
        return super_proxy != null && fail_count < max_failures;
    }

    public void close() {
        if (client != null)
            try { client.close(); } catch (IOException e) {}
        client = null;
    }
}

public class Example implements Runnable {
    public static final int n_parallel_exit_nodes = 100;
    public static final int n_total_req = 1000;
    public static final int switch_ip_every_n_req = 40;
    public static AtomicInteger at_req = new AtomicInteger(0);

    public static void main(String[] args) {
        ExecutorService executor =
            Executors.newFixedThreadPool(n_parallel_exit_nodes);
        for (int i = 0; i < n_parallel_exit_nodes; i++)
            executor.execute(new Example());
        executor.shutdown();
    }

    @Override
    public void run() {
        Client client = new Client(null);
        while (at_req.getAndAdd(1) < n_total_req) {
            if (!client.have_good_super_proxy())
                client.switch_session_id();
            if (client.n_req_for_exit_node == switch_ip_every_n_req)
                client.switch_session_id();
            CloseableHttpResponse response = null;
            try {
                response = client.request("http://lumtest.com/myip.json");
                int code = response.getStatusLine().getStatusCode();
                System.out.println(code != 200 ? code :
                        EntityUtils.toString(response.getEntity()));
            } catch (IOException e) {
                System.out.println(e.getMessage());
            } finally {
                try {
                    if (response != null)
                        response.close();
                } catch (Exception e) {}
            }
        }
        client.close();
    }
}

