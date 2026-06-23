---
name: pipeline-debug
description: Debug and fix GitLab CI/CD pipeline errors for Go, PHP, and Node.js projects
---

# Pipeline Debug Skill

Systematically debug and fix GitLab CI/CD pipeline errors.

## When to Use

- User pastes a pipeline error from GitLab Runner
- Pipeline fails with Docker, network, dependency, or build errors
- Need to identify root cause and apply fix

## Workflow

### 1. Parse the Error

Extract key information from the pipeline output:
- Runner type (shell, docker, etc.)
- Which stage failed (build, test, deploy)
- Error message (timeout, permission, network, syntax)

### 2. Read Configuration Files

Read these files to understand the pipeline setup:
```bash
cat .gitlab-ci.yml
cat Dockerfile
cat docker-compose.yml
cat package.json  # For Node.js
cat composer.json  # For PHP
cat go.mod  # For Go
```

### 3. Common Error Patterns

#### Network/Registry Errors
- **Symptoms**: `curl error 28`, `timeout`, `dial tcp: lookup failed`
- **Cause**: Server has restricted internet access
- **Fix**: Use ArvanCloud mirrors:
  - Docker Hub → `docker.arvancloud.ir/`
  - GitHub → proxy or mirror
  - Debian/Ubuntu → ArvanCloud apt mirror
  - Composer → Set `COMPOSER_HOME` with mirror config

#### Docker Build Errors
- **Symptoms**: `FROM` pull fails, layer download timeout
- **Fix**: Replace image references:
  ```yaml
  # Bad
  image: docker.io/library/golang:1.24
  
  # Good (for ArvanCloud)
  image: docker.arvancloud.ir/library/golang:1.24
  ```

#### Dependency Errors
- **PHP/Composer**: `Failed to download from dist: curl error 28`
  - Add `"preferred-install": "dist"` with fallback to source
  - Use `COMPOSER_NO_INTERACTION=1` for CI
  
- **Node.js/npm**: `npm ERR! Exit handler never called`
  - Usually Node version mismatch (old packages + Node 22)
  - Use `NODE_OPTIONS=--openssl-legacy-provider`

- **Go**: Module download timeout
  - Set `GOPROXY=https://goproxy.cn,direct` or mirror

#### Permission Errors
- **Symptoms**: `Permission denied`, `cannot create directory`
- **Fix**: Check `before_script` permissions, user context in Docker

### 4. Apply Fix

Edit the appropriate file:
- `.gitlab-ci.yml` - Pipeline configuration
- `Dockerfile` - Build environment
- `docker-compose.yml` - Service dependencies

### 5. Verify

After fix is applied, instruct user to re-run pipeline.

## Example Fixes

### ArvanCloud Docker Registry
```yaml
# Before
image: docker:latest

# After
image: docker.arvancloud.ir/library/docker:latest
```

### Composer with Mirror
```yaml
before_script:
  - export COMPOSER_HOME="$(pwd)/.composer"
  - echo '{"config":{"preferred-install":"dist","disable-tls":true}}' > "$COMPOSER_HOME/auth.json"
```

### Network-Limited apt-get
```dockerfile
# Replace default mirror
RUN sed -i 's|deb.debian.org|debian.arvancloud.ir|g' /etc/apt/sources.list
RUN apt-get update && apt-get install -y ...
```
