# docker-dns

Automatically create DNS entries for Traefik exposed docker containers

## Setup

A basic docker compose setup

```yaml
services:
  docker_dns:
    image: nicjohnson145/docker_dns:v0.1.0
    restart: unless-stopped
    environment:
      - "DOMAIN=nicjohnson.info"
      - "PROVIDER=cloudflare"
    volumes:
      - "/var/run/docker.sock:/var/run/docker.sock:ro"
      - "/path/to/credentials:/auth/credentials.ini"
```

Trigger docker dns to create an entry for a container by adding the label
`docker_dns.subdomain=<subdomain>`