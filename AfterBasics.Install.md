# After a basic install
Each utility provided by peertubestats has it's separate usage information and configuration options. See
 - [Usage of CronSaveStats](Usage%20of%20CronSaveStats.md)
 - [Usage of peertubeExportStat](Usage%20of%20peertubeExportStat.md)
 - [Usage of peertubestats](Usage%20of%20peertubestats.md)

After you confirmed that the service is up and running we suggest binding the service on localhost and exposing it via a reverse proxy like NGINX, as peertube stats does not provide ssl.
## A simple and working NGINX configuration (Live server)
```nginx
server {
    server_name peertubestats.example.com;
    location / {
        # Basic Authentication
        auth_basic "Protected Area";
        auth_basic_user_file /etc/nginx/.htpasswd;
        proxy_pass http://127.0.0.1:8080;
        proxy_request_buffering off;
        proxy_set_header x-forwarded-for $remote_addr;
    }

    listen 443 ssl; # managed by Certbot
    ssl_certificate /etc/letsencrypt/live/peertubestats.example.com/fullchain.pem; # managed by Certbot
    ssl_certificate_key /etc/letsencrypt/live/peertubestats.example.com/privkey.pem; # managed by Certbot
    include /etc/letsencrypt/options-ssl-nginx.conf; # managed by Certbot
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem; # managed by Certbot
}

server {
    if ($host = peertubestats.example.com) {
        return 301 https://$host$request_uri;
    } # managed by Certbot

    listen 80;
    server_name peertubestats.example.com;
    return 404; # managed by Certbot
}
```
This provides you with a password protected installation of the live service.
setting up the .htpasswd file [requires its own setup](https://docs.nginx.com/nginx/admin-guide/security-controls/configuring-http-basic-authentication/).

## A simple and working NGINX configuration (Static site)
```nginx
server {
    server_name peertubestats.example.com;
    location / {
        index index.html;
        root /opt/peertubestats/Reports/;
        # Basic Authentication
        auth_basic "Protected Area";
        auth_basic_user_file /etc/nginx/.htpasswd;
        proxy_pass http://127.0.0.1:8080;
        proxy_request_buffering off;
        proxy_set_header x-forwarded-for $remote_addr;
    }

    listen 443 ssl; # managed by Certbot
    ssl_certificate /etc/letsencrypt/live/peertubestats.example.com/fullchain.pem; # managed by Certbot
    ssl_certificate_key /etc/letsencrypt/live/peertubestats.example.com/privkey.pem; # managed by Certbot
    include /etc/letsencrypt/options-ssl-nginx.conf; # managed by Certbot
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem; # managed by Certbot
}

server {
    if ($host = peertubestats.example.com) {
        return 301 https://$host$request_uri;
    } # managed by Certbot

    listen 80;
    server_name peertubestats.example.com;
    return 404; # managed by Certbot
}
```

## Hosting recommendation.
peertubestats may be able to run without any additional software, however it does not handle ssl so a reverse proxy is recommended when serving to any other networks other than loopback.

peertubestats also does not provide any authentication, so all data is visible to anyone. For private peertube instances I recommend using nginx, caddy may also work but [SWAG](https://github.com/linuxserver/docker-swag) is a great option as well as it automates the administration of nginx. For anyone requiring authentication beyond basic auth I suggest checking out [authelia](https://github.com/authelia/authelia), which enables all kinds of authentication for your users.
