---
name: restricted-network-docker
description: Configure Docker, Composer, and package managers for ArvanCloud/restricted network environments
---

# Restricted Network Docker Skill

Configure Docker builds and package managers for servers with limited internet access (ArvanCloud).

## When to Use

- Server cannot access Docker Hub, GitHub, or public package registries
- Need to use ArvanCloud mirrors for all external dependencies
- Pipeline fails with network timeout errors

## ArvanCloud Mirrors

| Service | Original | ArvanCloud Mirror |
|---------|----------|-------------------|
| Docker Hub | `docker.io` | `docker.arvancloud.ir` |
| Debian/Ubuntu | `deb.debian.org` | `debian.arvancloud.ir` |
| Alpine | `dl-cdn.alpinelinux.org` | `dl-cdn.alpinelinux.org.arvancloud.ir` |
| GitHub | `github.com` | Use proxy or local mirror |
| npm | `registry.npmjs.org` | `registry.npmjs.org.arvancloud.ir` |
| PyPI | `pypi.org` | `pypi.arvancloud.ir` |

## Dockerfile Fixes

### Replace Base Images

```dockerfile
# Before (will timeout)
FROM golang:1.24-alpine
FROM node:22-alpine
FROM php:8.2-fpm
FROM composer:latest

# After (ArvanCloud)
FROM docker.arvancloud.ir/library/golang:1.24-alpine
FROM docker.arvancloud.ir/library/node:22-alpine
FROM docker.arvancloud.ir/library/php:8.2-fpm
FROM docker.arvancloud.ir/library/composer:latest
```

### Debian/Ubuntu apt Mirror

```dockerfile
# Replace default mirror
RUN sed -i 's|deb.debian.org|debian.arvancloud.ir|g' /etc/apt/sources.list && \
    apt-get update && \
    apt-get install -y --no-install-recommends \
    package1 package2

# Clean up
RUN apt-get clean && rm -rf /var/lib/apt/lists/*
```

### Alpine apk Mirror

```dockerfile
# Replace default mirror
RUN sed -i 's|https://dl-cdn.alpinelinux.org|https://dl-cdn.alpinelinux.org.arvancloud.ir|g' /etc/apk/repositories && \
    apk add --no-cache package1 package2
```

## Composer Fixes

### composer.json Configuration

```json
{
  "config": {
    "preferred-install": "dist",
    "disable-tls": true,
    "github-protocols": ["http", "git"],
    "process-timeout": 300
  },
  "repositories": [
    {
      "type": "composer",
      "url": "https://packagist.arvancloud.ir"
    }
  ]
}
```

### CI/CD Environment Variables

```yaml
before_script:
  - export COMPOSER_HOME="$(pwd)/.composer"
  - mkdir -p "$COMPOSER_HOME"
  - |
    cat > "$COMPOSER_HOME/auth.json" << 'EOF'
    {
      "config": {
        "preferred-install": "dist",
        "disable-tls": true
      },
      "repositories": [
        {
          "type": "composer",
          "url": "https://packagist.arvancloud.ir"
        }
      ]
    }
    EOF
  - composer install --no-interaction --prefer-dist
```

## npm/Node.js Fixes

### .npmrc Configuration

```ini
registry=https://registry.npmjs.org.arvancloud.ir
strict-ssl=false
```

### CI/CD Setup

```yaml
before_script:
  - npm config set registry https://registry.npmjs.org.arvancloud.ir
  - npm install --prefer-offline
```

## Go Module Fixes

```bash
# Set Go proxy
export GOPROXY=https://goproxy.cn,https://goproxy.io,direct
go mod download
```

## GitLab CI/CD Docker-in-Docker

```yaml
build:
  image: docker.arvancloud.ir/library/docker:latest
  services:
    - docker.arvancloud.ir/library/docker:24-dind
  variables:
    DOCKER_TLS_CERTDIR: ""
  before_script:
    - docker login -u "$HARBOR_USER" -p "$HARBOR_PASS" harbor.sepehritg.ir
  script:
    - docker build -t harbor.sepehritg.ir/project:latest .
    - docker push harbor.sepehritg.ir/project:latest
```

## Debugging Network Issues

```bash
# Test connectivity
docker run --rm docker.arvancloud.ir/library/alpine:latest echo "OK"

# Check DNS
nslookup docker.arvancloud.ir

# Test specific port
curl -v https://docker.arvancloud.ir/v2/

# Check proxy settings
echo $HTTP_PROXY $HTTPS_PROXY $NO_PROXY
```
