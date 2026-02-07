# Deployment Guide

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Kubernetes Deployment](#kubernetes-deployment)
3. [Docker Swarm](#docker-swarm)
4. [Standalone Server](#standalone-server)
5. [Production Hardening](#production-hardening)
6. [Monitoring and Alerting](#monitoring-and-alerting)

## Prerequisites

- WireGuard installed and configured on the host/network
- API key generated (use `openssl rand -base64 32`)
- HTTPS certificate and key (for TLS)

## Kubernetes Deployment

### 1. Create a Secret for the API Key

```bash
kubectl create secret generic wg-agent-api-key \
  --from-literal=api-key=$(openssl rand -base64 32)
```

### 2. Create ConfigMap

```bash
kubectl create configmap wg-agent-config \
  --from-literal=log-level=info \
  --from-literal=wg-interface=wg0
```

### 3. Create the Deployment

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: wg-agent
  namespace: wireguard

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: wg-agent
  namespace: wireguard
spec:
  replicas: 1
  selector:
    matchLabels:
      app: wg-agent
  template:
    metadata:
      labels:
        app: wg-agent
    spec:
      serviceAccountName: wg-agent
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 1000
      containers:
      - name: wg-agent
        image: ghcr.io/afrisinc/wg-agent:latest
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 9999
          name: http
          protocol: TCP
        env:
        - name: API_KEY
          valueFrom:
            secretKeyRef:
              name: wg-agent-api-key
              key: api-key
        - name: HOST
          value: "0.0.0.0"
        - name: PORT
          value: "9999"
        - name: WG_INTERFACE
          value: "wg0"
        - name: LOG_LEVEL
          valueFrom:
            configMapKeyRef:
              name: wg-agent-config
              key: log-level
        resources:
          requests:
            memory: "64Mi"
            cpu: "100m"
          limits:
            memory: "128Mi"
            cpu: "500m"
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          capabilities:
            add:
            - NET_ADMIN
            - SYS_MODULE
        livenessProbe:
          httpGet:
            path: /health
            port: 9999
          initialDelaySeconds: 10
          periodSeconds: 30
          timeoutSeconds: 5
        readinessProbe:
          httpGet:
            path: /ready
            port: 9999
          initialDelaySeconds: 5
          periodSeconds: 10
          timeoutSeconds: 5
        volumeMounts:
        - name: wireguard
          mountPath: /etc/wireguard
          readOnly: true
        - name: tmp
          mountPath: /tmp
      volumes:
      - name: wireguard
        hostPath:
          path: /etc/wireguard
          type: Directory
      - name: tmp
        emptyDir: {}
      nodeSelector:
        wireguard: enabled

---
apiVersion: v1
kind: Service
metadata:
  name: wg-agent
  namespace: wireguard
spec:
  type: ClusterIP
  ports:
  - port: 9999
    targetPort: 9999
    protocol: TCP
    name: http
  selector:
    app: wg-agent

---
apiVersion: autoscaling.k8s.io/v2
kind: HorizontalPodAutoscaler
metadata:
  name: wg-agent
  namespace: wireguard
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: wg-agent
  minReplicas: 1
  maxReplicas: 3
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 80
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 85
```

### 4. Deploy

```bash
kubectl apply -f deployment.yaml
```

### 5. Verify

```bash
kubectl get pods -n wireguard
kubectl logs -n wireguard -f deployment/wg-agent
```

## Docker Swarm

### 1. Create Secret

```bash
echo "your-api-key" | docker secret create wg-agent-key -
```

### 2. Deploy Service

```bash
docker service create \
  --name wg-agent \
  --secret wg-agent-key \
  --publish 9999:9999 \
  --cap-add NET_ADMIN \
  --cap-add SYS_MODULE \
  --mount type=bind,source=/etc/wireguard,target=/etc/wireguard,readonly \
  --env API_KEY_FILE=/run/secrets/wg-agent-key \
  --env HOST=0.0.0.0 \
  --env PORT=9999 \
  --env WG_INTERFACE=wg0 \
  --update-parallelism 1 \
  --update-delay 10s \
  --health-cmd="wget --quiet --tries=1 --spider http://localhost:9999/health || exit 1" \
  --health-interval=30s \
  --health-timeout=5s \
  --health-retries=3 \
  ghcr.io/afrisinc/wg-agent:latest
```

## Standalone Server

### 1. Prerequisites

```bash
sudo apt-get update
sudo apt-get install -y wireguard-tools ca-certificates
```

### 2. Create User and Directory

```bash
sudo useradd -r -s /bin/false wg-agent
sudo mkdir -p /opt/wg-agent
sudo chown -R wg-agent:wg-agent /opt/wg-agent
```

### 3. Download Binary

```bash
VERSION=v1.0.0
sudo wget -O /opt/wg-agent/wg-agent \
  https://github.com/afrisinc/wg-agent/releases/download/${VERSION}/wg-agent-linux-amd64
sudo chmod +x /opt/wg-agent/wg-agent
```

### 4. Create Systemd Service

Create `/etc/systemd/system/wg-agent.service`:

```ini
[Unit]
Description=WireGuard Agent
Documentation=https://github.com/afrisinc/wg-agent
After=wireguard.service
Wants=wireguard.service

[Service]
Type=simple
User=wg-agent
Group=wg-agent
ExecStart=/opt/wg-agent/wg-agent

# Security settings
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=/etc/wireguard

# Capabilities
AmbientCapabilities=CAP_NET_ADMIN CAP_SYS_MODULE
CapabilityBoundingSet=CAP_NET_ADMIN CAP_SYS_MODULE CAP_SETUID CAP_SETGID

# Environment
Environment="API_KEY=your-api-key"
Environment="HOST=0.0.0.0"
Environment="PORT=9999"
Environment="WG_INTERFACE=wg0"
Environment="LOG_LEVEL=info"

# Restart policy
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### 5. Start Service

```bash
sudo systemctl daemon-reload
sudo systemctl enable wg-agent
sudo systemctl start wg-agent
```

### 6. Verify

```bash
sudo systemctl status wg-agent
sudo journalctl -u wg-agent -f
```

## Production Hardening

### 1. TLS/HTTPS

Use a reverse proxy like Nginx:

```nginx
upstream wg_agent {
    server 127.0.0.1:9999;
}

server {
    listen 443 ssl http2;
    server_name wg-agent.example.com;

    ssl_certificate /etc/letsencrypt/live/wg-agent.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/wg-agent.example.com/privkey.pem;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-Frame-Options "DENY" always;
    add_header X-XSS-Protection "1; mode=block" always;

    location / {
        proxy_pass http://wg_agent;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Timeouts
        proxy_connect_timeout 10s;
        proxy_send_timeout 10s;
        proxy_read_timeout 10s;
    }
}

server {
    listen 80;
    server_name wg-agent.example.com;
    return 301 https://$server_name$request_uri;
}
```

### 2. Rate Limiting

Add to Nginx configuration:

```nginx
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;
limit_req zone=api_limit burst=20 nodelay;
```

### 3. API Key Rotation

1. Generate new key: `openssl rand -base64 32`
2. Update secret: `kubectl patch secret wg-agent-api-key -p '{"data":{"api-key":"<NEW_KEY_BASE64>"}}'`
3. Restart pods: `kubectl rollout restart deployment/wg-agent`

### 4. Firewall Rules

```bash
# Allow only from trusted sources
sudo ufw allow from 10.0.0.0/8 to any port 9999
sudo ufw allow from 192.168.0.0/16 to any port 9999
```

## Monitoring and Alerting

### Prometheus Metrics

Create a monitor scrape configuration:

```yaml
- job_name: 'wg-agent'
  static_configs:
    - targets: ['localhost:9999']
  metrics_path: '/metrics'
```

### Alerting Rules

```yaml
groups:
- name: wg-agent
  rules:
  - alert: WGAgentDown
    expr: up{job="wg-agent"} == 0
    for: 5m
    annotations:
      summary: "WireGuard Agent is down"

  - alert: WGAgentHighErrorRate
    expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
    for: 5m
    annotations:
      summary: "WireGuard Agent high error rate"

  - alert: WGAgentUnhealthy
    expr: up{job="wg-agent"} == 1 AND http_requests_failed_total > 10
    for: 10m
    annotations:
      summary: "WireGuard Agent unhealthy"
```

### Logging

Use centralized logging (ELK, Loki, etc.):

```bash
# Example with Loki labels
- job_name: wg-agent
  pipeline_stages:
    - json:
        expressions:
          level: level
          msg: msg
  relabel_configs:
    - source_labels: ['__hostname__']
      target_label: hostname
```

## Rollback Procedure

1. **Identify issue**: Check logs and metrics
2. **Scale down new version**: `kubectl scale deployment wg-agent --replicas=0`
3. **Deploy previous version**: `kubectl set image deployment/wg-agent wg-agent=ghcr.io/afrisinc/wg-agent:v1.0.0`
4. **Verify**: Check health probes and logs
5. **Scale up**: `kubectl scale deployment wg-agent --replicas=3`

## Health Check

```bash
# Health endpoint
curl http://localhost:9999/health

# Readiness endpoint
curl http://localhost:9999/ready

# Test API endpoint
curl -H "X-API-Key: your-api-key" http://localhost:9999/public-key
```

## Support

For deployment issues, see the [TROUBLESHOOTING.md](TROUBLESHOOTING.md) guide.
