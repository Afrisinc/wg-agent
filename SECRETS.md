# GitHub Secrets Configuration

To enable CI/CD and automated deployments, configure these secrets in your GitHub repository.

## How to Add Secrets

1. Go to: **Settings** → **Secrets and variables** → **Actions**
2. Click **New repository secret**
3. Add each secret below

## Required Secrets

### Docker Registry Credentials (GHCR)

These are used to push Docker images to GitHub Container Registry.

| Secret | Description | Example |
|--------|-------------|---------|
| `GHCR_OWNER` | GitHub username or org | `afrisinc` |
| `GHCR_TOKEN` | GitHub personal access token | `ghp_xxxxxxxxxxxx` |

**How to create GHCR_TOKEN:**
1. Go to: https://github.com/settings/tokens
2. Click **Generate new token** → **Generate new token (classic)**
3. Select scopes:
   - `write:packages` - to push packages
   - `read:packages` - to pull packages
   - `delete:packages` - to delete packages
4. Copy the token and add as `GHCR_TOKEN` secret

### VPS Deployment Credentials

| Secret | Description | Example |
|--------|-------------|---------|
| `VPS_HOST` | VPS IP or hostname | `123.45.67.89` |
| `VPS_USER` | SSH username | `root` or `deploy` |
| `VPS_SSH_KEY` | Private SSH key (multiline) | `-----BEGIN RSA PRIVATE KEY-----...` |
| `VPS_PORT` | SSH port (optional) | `22` |
| `VPS_APP_PATH` | App directory on VPS | `/opt/apps/wg-agent` |

**How to generate SSH key:**
```bash
# Generate key (leave passphrase empty for CI/CD)
ssh-keygen -t rsa -b 4096 -f ~/.ssh/id_rsa_ci -N ""

# Copy the private key
cat ~/.ssh/id_rsa_ci

# Add public key to VPS authorized_keys
ssh-copy-id -i ~/.ssh/id_rsa_ci.pub user@vps-host
```

Then:
1. Copy the entire private key content
2. Add as `VPS_SSH_KEY` secret (include the `-----BEGIN` and `-----END` lines)

### Application Secrets

| Secret | Description | Example |
|--------|-------------|---------|
| `WG_AGENT_API_KEY` | WireGuard Agent API key | `kJ7hL+M9oP2qR3sT4uV5wX6yZ7aB8cD9=` |
| `WG_AGENT_HOST` | Host to bind to (optional) | `0.0.0.0` |
| `WG_AGENT_PORT` | Port to use (optional) | `9999` |
| `DOCKER_REGISTRY` | Custom Docker registry (optional) | `docker.io/youruser` |

**How to generate API_KEY:**
```bash
# On your machine
openssl rand -base64 32
```

### Notification Secrets (Optional)

#### Slack Notifications

| Secret | Description |
|--------|-------------|
| `SLACK_WEBHOOK` | Slack incoming webhook URL |

**How to create:**
1. Go to: https://api.slack.com/apps
2. Create new app → "From scratch"
3. Enable Incoming Webhooks
4. Add New Webhook to Workspace
5. Copy the Webhook URL

#### Discord Notifications

| Secret | Description |
|--------|-------------|
| `DISCORD_WEBHOOK` | Discord webhook URL |

**How to create:**
1. Open Discord server
2. Right-click channel → Edit channel
3. Integrations → Webhooks → New Webhook
4. Copy the webhook URL

## Security Best Practices

### SSH Key Security

❌ **DO NOT:**
- Commit SSH keys to the repository
- Use keys with passphrases in CI/CD
- Share SSH keys

✅ **DO:**
- Generate dedicated CI/CD keys
- Use deployment user with minimal permissions
- Rotate keys periodically
- Monitor key usage in logs

### API Key Security

❌ **DO NOT:**
- Use the same API key in multiple places
- Log or expose the API key
- Commit `.env` files with secrets

✅ **DO:**
- Use strong, randomly generated keys
- Rotate keys quarterly
- Use secrets management (Vault, 1Password)
- Monitor API usage

### SSH User Permissions

Create a dedicated deploy user on your VPS:

```bash
# On VPS as root
useradd -m -s /bin/bash deploy
usermod -aG docker deploy

# Add public key
mkdir -p /home/deploy/.ssh
echo "$(cat ~/.ssh/id_rsa_ci.pub)" >> /home/deploy/.ssh/authorized_keys
chmod 600 /home/deploy/.ssh/authorized_keys
chown deploy:deploy /home/deploy/.ssh -R

# Test SSH login
ssh -i ~/.ssh/id_rsa_ci deploy@vps-host
```

Then use `deploy` as `VPS_USER` instead of `root`.

## Testing Deployment Locally

Before pushing to GitHub, test the deployment script:

```bash
# Set environment variables
export VPS_HOST="your-vps-ip"
export VPS_USER="deploy"
export VPS_SSH_KEY="$(cat ~/.ssh/id_rsa_ci)"
export VPS_PORT="22"
export VPS_APP_PATH="/opt/apps/wg-agent"
export DEPLOY_GHCR_USERNAME="your-username"
export DEPLOY_GHCR_TOKEN="your-token"
export DOCKER_IMAGE="ghcr.io/your-org/wg-agent"
export BRANCH_NAME="main"
export API_KEY="your-api-key"

# Test SSH connection
ssh -i ~/.ssh/id_rsa_ci -p $VPS_PORT $VPS_USER@$VPS_HOST "echo 'SSH works!'"
```

## Secrets Checklist

### Minimal Deployment (GitHub Container Registry + VPS)

- [ ] `GHCR_OWNER` - GitHub username
- [ ] `GHCR_TOKEN` - GitHub token with package write permissions
- [ ] `VPS_HOST` - VPS IP or hostname
- [ ] `VPS_USER` - SSH user (e.g., `deploy`)
- [ ] `VPS_SSH_KEY` - Private SSH key
- [ ] `VPS_APP_PATH` - App directory (default: `/opt/apps/wg-agent`)
- [ ] `WG_AGENT_API_KEY` - WireGuard Agent API key

### Full Deployment (Above + Notifications)

- [ ] All minimal secrets
- [ ] `SLACK_WEBHOOK` - Slack webhook (optional)
- [ ] `DISCORD_WEBHOOK` - Discord webhook (optional)

### Advanced Deployment (All + Custom Registry)

- [ ] All full deployment secrets
- [ ] `DOCKER_REGISTRY` - Custom Docker registry
- [ ] `DEPLOY_GHCR_USERNAME` - Custom registry username
- [ ] `DEPLOY_GHCR_TOKEN` - Custom registry token

## Troubleshooting

### Deployment fails: "SSH authentication failed"

1. Verify SSH key is added to `VPS_SSH_KEY` secret (multiline)
2. Check public key is in VPS `~/.ssh/authorized_keys`
3. Verify VPS_USER has SSH access
4. Test manually:
   ```bash
   ssh -i private-key -p VPS_PORT VPS_USER@VPS_HOST "whoami"
   ```

### Docker login fails: "invalid username or password"

1. Verify `GHCR_TOKEN` has correct permissions
2. Check token hasn't expired (30 days default)
3. Verify username matches `GHCR_OWNER`
4. Test manually:
   ```bash
   echo "TOKEN" | docker login ghcr.io -u USERNAME --password-stdin
   ```

### Image pull fails on VPS: "image not found"

1. Verify image was pushed successfully (check GitHub Packages)
2. Check `VPS_USER` can pull from GHCR
3. Verify credentials in deploy script
4. Check image name format: `ghcr.io/owner/repo:tag`

### Health check fails after deployment

1. Check container logs:
   ```bash
   docker compose logs -f wg-agent
   ```
2. Verify WireGuard is installed on VPS:
   ```bash
   which wg wg-quick
   ```
3. Verify API_KEY is set correctly
4. Check port binding:
   ```bash
   netstat -tlnp | grep 9999
   ```

## More Information

- GitHub Secrets: https://docs.github.com/en/actions/security-guides/encrypted-secrets
- SSH Keys: https://docs.github.com/en/authentication/connecting-to-github-with-ssh
- GitHub Container Registry: https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry
