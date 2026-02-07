# Troubleshooting Guide

## Common Issues

### 1. "Unauthorized" Error

**Symptom**: Getting 401 Unauthorized responses

**Solution**:
- Verify API key is set: `echo $API_KEY`
- Check API key in request header: `-H "X-API-Key: your-api-key"`
- Ensure API key matches exactly (case-sensitive)
- Verify header name is correct: `X-API-Key` (not `x-api-key`)

```bash
# Test with correct header
curl -H "X-API-Key: $(echo $API_KEY)" http://localhost:9999/health
```

### 2. "Connection Refused"

**Symptom**: Cannot connect to the agent

**Solution**:
- Check if agent is running: `ps aux | grep wg-agent`
- Verify port is listening: `netstat -tlnp | grep 9999`
- Check firewall: `sudo ufw status`
- Verify host/port settings: `echo $HOST:$PORT`

```bash
# Start the agent
go run main.go

# Check logs
journalctl -u wg-agent -f
```

### 3. "WireGuard Interface Not Found"

**Symptom**: Getting "interface not found" or similar errors

**Solution**:
- Verify WireGuard is installed: `which wg`
- Check interface exists: `wg show wg0`
- Verify interface name: `ip link show`
- Check WG_INTERFACE env var: `echo $WG_INTERFACE`

```bash
# Bring up the interface
sudo wg-quick up wg0

# Verify
wg show
```

### 4. "Permission Denied" Errors

**Symptom**: Getting permission denied when running commands

**Solution**:
- Ensure agent has CAP_NET_ADMIN capability
- In Docker: `--cap-add NET_ADMIN --cap-add SYS_MODULE`
- In Kubernetes: Add capabilities in security context
- Run as root or with sudo (not recommended in production)

```bash
# Check capabilities
getcap /path/to/wg-agent

# Or run with capabilities
sudo setcap cap_net_admin=ep /path/to/wg-agent
```

### 5. Invalid Public Key Error

**Symptom**: "Public key format is invalid"

**Solution**:
- WireGuard public keys are base64-encoded and end with "="
- Public key format: 44 characters (43 + 1 equals sign)
- Valid example: `jI6DsucHvzJzcow3v7CqvJODct9+pWG8V+MlaWL7yGc=`

```bash
# Generate a valid public/private key pair
wg genkey | tee privatekey | wg pubkey > publickey
```

### 6. Invalid IP Address Error

**Symptom**: "IP address format is invalid"

**Solution**:
- Use standard IPv4 format: `XXX.XXX.XXX.XXX`
- Valid examples: `192.168.1.100`, `10.0.0.5`
- Invalid: `192.168.1`, `192.168.1.256`, `192.168.1.1/32`

```bash
# Correct format (without CIDR)
curl -H "X-API-Key: $API_KEY" \
  -d '{"public_key": "...", "allowed_ip": "192.168.1.100"}' \
  http://localhost:9999/add-peer
```

### 7. "Failed to Save Config"

**Symptom**: Peer added but configuration failed to save

**Solution**:
- Verify `/etc/wireguard/wg0.conf` is writable
- Check disk space: `df -h`
- Verify wg-quick is installed: `which wg-quick`
- Check file permissions: `ls -la /etc/wireguard/`

```bash
# Verify file exists and is writable
sudo touch /etc/wireguard/wg0.conf
sudo chmod 600 /etc/wireguard/wg0.conf
sudo chown root:root /etc/wireguard/wg0.conf
```

### 8. Peer Addition Succeeds but Peer Not Active

**Symptom**: Peer was added but doesn't appear in `wg show`

**Solution**:
- Run `wg show` to verify peer exists
- Check WireGuard interface is up: `ip link show wg0`
- Verify public key is correct
- Check interface has an address: `ip addr show wg0`

```bash
# Verify interface is up
sudo ip link set wg0 up

# Assign an address if needed
sudo ip addr add 10.0.0.1/24 dev wg0

# Bring up with wg-quick
sudo wg-quick up wg0
```

### 9. Docker Container Keeps Exiting

**Symptom**: Container exits immediately

**Solution**:
- Check logs: `docker logs wg-agent`
- Verify environment variables: `docker inspect wg-agent`
- Ensure API_KEY is set
- Check WireGuard is available on host: `wg show`
- Verify required capabilities are passed

```bash
# Run with proper options
docker run -it --rm \
  --cap-add NET_ADMIN \
  --cap-add SYS_MODULE \
  -v /etc/wireguard:/etc/wireguard:ro \
  -e API_KEY=test-key \
  -e WG_INTERFACE=wg0 \
  wg-agent:latest
```

### 10. High CPU or Memory Usage

**Symptom**: Agent consuming excessive resources

**Solution**:
- Check for command timeouts: Verify timeout values in code
- Review logs for stuck operations
- Limit concurrent connections at proxy level
- Implement rate limiting

```bash
# Monitor resource usage
docker stats wg-agent

# Set resource limits
docker run --memory 128m --cpus 0.5 wg-agent:latest
```

## Debugging

### Enable Verbose Logging

```bash
# Set log level
export LOG_LEVEL=debug
go run main.go
```

### Test API Manually

```bash
# Test health endpoint (no auth needed)
curl -v http://localhost:9999/health

# Test authenticated endpoint
curl -v \
  -H "X-API-Key: test-key" \
  http://localhost:9999/public-key

# Test add-peer with payload
curl -v \
  -H "X-API-Key: test-key" \
  -H "Content-Type: application/json" \
  -d '{
    "public_key": "jI6DsucHvzJzcow3v7CqvJODct9+pWG8V+MlaWL7yGc=",
    "allowed_ip": "192.168.1.100"
  }' \
  http://localhost:9999/add-peer
```

### Check System State

```bash
# Verify WireGuard installation
which wg wg-quick

# Check interface
ip link show
wg show

# Check running processes
ps aux | grep -E 'wg|agent'

# Check listening ports
netstat -tlnp | grep 9999

# Check system logs
journalctl -xe
dmesg | tail -20
```

### Test with Docker

```bash
# Interactive shell
docker run -it --rm \
  --cap-add NET_ADMIN \
  -v /etc/wireguard:/etc/wireguard:ro \
  -e API_KEY=test \
  wg-agent:latest /bin/sh

# Check inside container
apk add curl wget
wget http://localhost:9999/health
```

## Performance Tuning

### Optimize for High Throughput

1. Increase file descriptors:
   ```bash
   ulimit -n 65536
   ```

2. Tune kernel parameters:
   ```bash
   sysctl -w net.ipv4.netfilter.ip_conntrack_max=1000000
   sysctl -w net.ipv4.ip_local_port_range="1024 65535"
   ```

3. In Docker/Kubernetes, set resource requests:
   ```yaml
   resources:
     requests:
       memory: "64Mi"
       cpu: "100m"
     limits:
       memory: "256Mi"
       cpu: "1000m"
   ```

## Getting Help

1. Check logs: `journalctl -u wg-agent -f`
2. Review this guide
3. Check GitHub issues: https://github.com/afrisinc/wg-agent/issues
4. Collect diagnostic info:

```bash
# Diagnostic bundle
mkdir -p diagnostics
echo "=== Agent Status ===" >> diagnostics/info.txt
systemctl status wg-agent >> diagnostics/info.txt 2>&1
echo "=== WireGuard Info ===" >> diagnostics/info.txt
wg show >> diagnostics/info.txt 2>&1
echo "=== System Info ===" >> diagnostics/info.txt
uname -a >> diagnostics/info.txt
echo "=== Docker Info ===" >> diagnostics/info.txt
docker ps -a >> diagnostics/info.txt 2>&1
echo "=== Recent Logs ===" >> diagnostics/info.txt
journalctl -u wg-agent -n 100 >> diagnostics/info.txt 2>&1

# Create tarball
tar -czf diagnostics.tar.gz diagnostics/
```

This diagnostic bundle can be attached to GitHub issues for faster debugging.
