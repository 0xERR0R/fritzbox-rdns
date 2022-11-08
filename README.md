# Reverse DNS server for Fritzbox

This is a DNS server which fetches periodically network device information (Device name and IPv4+IPv6 addresses) from the Fritzbox web ui and provides device name for reverse DNS (PTR) queries. Works for IPv4 and IPv6 addresses. This tool can be configured als rDNS lookup in [blocky](https://github.com/0xERR0R/blocky) or AdguardHome.

## How to install

You can start this server as docker container with following `docker-compose.yaml` file

```yaml
version: "2.1"
services:
  fritzbox-rdns:
    container_name: fritzbox-rdns
    image: ghcr.io/0xerr0r/fritzbox-rdns:latest
    restart: unless-stopped
    environment:
      - "FB_URL=http://192.168.178.1"
      - "FB_USER=username"
      - "FB_PASSWORD=passw0rd"
      - "FB_REDIS=redis:6379"
      - "FB_LOG_LEVEL=info"
    ports:
      - "53:53/udp"
  redis:
    image: redis
    restart: unless-stopped
    command: redis-server --save 120 1 --loglevel warning
    volumes:
      - redis_data:/data
volumes:
    redis_data:
```

## How to test

To resolve the hostname from 192.168.178.3, please execute (change host and port accordingly)

```sh
dig @host -p 53 -x 192.168.178.3
```

This should return something like:

```
;; ANSWER SECTION:
3.178.168.192.in-addr.arpa. 300 IN      PTR     laptop.
```