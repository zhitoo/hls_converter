---
name: k8s-deploy
description: Set up Kubernetes deployment with GitLab CI/CD for Go/Node.js/PHP services
---

# Kubernetes Deploy Skill

Create complete Kubernetes deployment setup with GitLab CI/CD pipeline.

## When to Use

- User wants to deploy a service to Kubernetes
- Need to create kube.yml, Dockerfile, and CI/CD pipeline
- Setting up production deployment with resource limits

## Workflow

### 1. Understand the Service

Identify:
- Language/framework (Go, Node.js, PHP)
- Port the service listens on
- Dependencies (database, volumes, external services)
- Environment variables needed

### 2. Create/Update Dockerfile

```dockerfile
# Multi-stage build for Go
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/main .
RUN addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -s /bin/sh -D appuser
USER appuser
EXPOSE 8080
CMD ["./main"]
```

### 3. Create kube.yml

Template with these sections:
- **Namespace** (optional, for isolation)
- **Deployment** with resource limits
- **Service** (ClusterIP or LoadBalancer)
- **Ingress** (if using ingress controller)
- **ConfigMap** (for non-sensitive config)
- **Secret** (for credentials - created manually)

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: SERVICE_NAME
  labels:
    app: SERVICE_NAME
spec:
  replicas: 1
  selector:
    matchLabels:
      app: SERVICE_NAME
  template:
    metadata:
      labels:
        app: SERVICE_NAME
    spec:
      containers:
        - name: SERVICE_NAME
          image: HARBOR_REGISTRY/SERVICE_NAME:latest
          ports:
            - containerPort: 8080
          resources:
            requests:
              memory: "512Mi"
              cpu: "500m"
            limits:
              memory: "1Gi"
              cpu: "1"
          volumeMounts:
            - name: config
              mountPath: /app/config
              readOnly: true
      volumes:
        - name: config
          secret:
            secretName: SERVICE_NAME-config
---
apiVersion: v1
kind: Service
metadata:
  name: SERVICE_NAME
spec:
  selector:
    app: SERVICE_NAME
  ports:
    - port: 80
      targetPort: 8080
  type: ClusterIP
```

### 4. Create .gitlab-ci.yml

```yaml
stages:
  - build
  - deploy

variables:
  DOCKER_TLS_CERTDIR: ""
  HARBOR_REGISTRY: harbor.sepehritg.ir
  K8S_NAMESPACE: default

build:
  stage: build
  image: docker.arvancloud.ir/library/docker:latest
  services:
    - docker.arvancloud.ir/library/docker:24-dind
  before_script:
    - docker login -u "$HARBOR_USER" -p "$HARBOR_PASS" "$HARBOR_REGISTRY"
  script:
    - docker build -t "$HARBOR_REGISTRY/$CI_PROJECT_NAME:$CI_COMMIT_SHA" .
    - docker push "$HARBOR_REGISTRY/$CI_PROJECT_NAME:$CI_COMMIT_SHA"
  only:
    - main

deploy:
  stage: deploy
  image: bitnami/kubectl:latest
  before_script:
    - mkdir -p ~/.ssh && echo "$K8S_SSH_KEY" > ~/.ssh/id_rsa && chmod 600 ~/.ssh/id_rsa
  script:
    - ssh -o StrictHostKeyChecking=no root@K8S_SERVER_IP "kubectl set image deployment/$CI_PROJECT_NAME $CI_PROJECT_NAME=$HARBOR_REGISTRY/$CI_PROJECT_NAME:$CI_COMMIT_SHA -n $K8S_NAMESPACE"
  only:
    - main
  when: manual
```

### 5. Document Manual Steps

User must create secrets manually:
```bash
kubectl create secret generic SERVICE_NAME-config \
  --from-literal=KEY1=value1 \
  --from-file=config.json=./config.json
```

## Resource Limits Guidance

| Service Type | CPU Request | CPU Limit | Memory Request | Memory Limit |
|--------------|-------------|-----------|----------------|--------------|
| Go API | 500m | 2 | 512Mi | 2Gi |
| Node.js | 500m | 2 | 512Mi | 2Gi |
| PHP-FPM | 500m | 1 | 256Mi | 1Gi |
| Static | 100m | 500m | 64Mi | 256Mi |

## ArvanCloud Considerations

- Docker images must be pulled from `docker.arvancloud.ir`
- Harbor registry for private images
- SSH access for deployment (not kubectl directly from CI)
- Use `--validate=false` for kubectl apply in restricted networks
