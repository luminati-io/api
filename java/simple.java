package example;

import java.io.*;
import java.util.Random;
import org.apache.http.HttpHost;
import org.apache.http.auth.*;
import org.apache.http.client.CredentialsProvider;
import org.apache.http.client.fluent.Request;
import org.apache.http.client.methods.*;
import org.apache.http.impl.client.*;
import org.apache.http.impl.conn.BasicHttpClientConnectionManager;
import org.apache.http.util.EntityUtils;

class Client {
    public static final String username = "lum-customer-CUSTOMER-zone-YOURZONE";
    public static final String password = "YOURPASS";
    public static final int port = 22225;
    public String session_id = Integer.toString(new Random().nextInt(Integer.MAX_VALUE));
    public CloseableHttpClient client;

    public Client(String country) {
        String login = username+(country!=null ? "-country-"+country : "")
            +"-session-" + session_id;
        HttpHost super_proxy = new HttpHost("zproxy.luminati.io", port);
        CredentialsProvider cred_provider = new BasicCredentialsProvider();
        cred_provider.setCredentials(new AuthScope(super_proxy),
            new UsernamePasswordCredentials(login, password));
        client = HttpClients.custom()
            .setConnectionManager(new BasicHttpClientConnectionManager())
            .setProxy(super_proxy)
            .setDefaultCredentialsProvider(cred_provider)
            .build();
    }

    public String request(String url) throws IOException {
        HttpGet request = new HttpGet(url);
        CloseableHttpResponse response = client.execute(request);
        try {
            return EntityUtils.toString(response.getEntity());
        } finally { response.close(); }
    }

    public void close() throws IOException { client.close(); }
}

public class Example {
    public static void main(String[] args) throws IOException {
        System.out.println("Performing request(s)");
        Client client = new Client(null);
        try {
            // Put complete scraping sequence below:
            System.out.println(client.request("http://lumtest.com/myip.json"));
            // System.out.println(client.request(...second request...));
        } finally { client.close(); }
    }
}
