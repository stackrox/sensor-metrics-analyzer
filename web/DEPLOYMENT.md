# Deployment Guide

This guide covers deploying the Sensor Metrics Analyzer Web service on a Linux server.

## Prerequisites

- Linux server with systemd
- Nginx installed
- Go 1.21+ (for building, or use pre-built binaries)
- Root or sudo access

## Step 1: Prepare the Application

1. Clone or copy the repository to the server:
   ```bash
   cd /opt
   git clone <repository-url> sensor-metrics-analyzer
   # Or copy the files to /opt/sensor-metrics-analyzer
   ```

2. Download the precompiled binary from GitHub Releases:
   ```bash
   cd /opt/sensor-metrics-analyzer/sensor-metrics-analyzer-go
   curl -L -o bin/web-server https://github.com/stackrox/sensor-metrics-analyzer/releases/latest/download/web-server-linux-amd64
   chmod +x bin/web-server
   ```

3. Verify the binary exists:
   ```bash
   ls -lh /opt/sensor-metrics-analyzer/sensor-metrics-analyzer-go/bin/
   # Should show: web-server
   ```

## Step 2: Create System User

Create a dedicated user for running the service:

```bash
sudo useradd -r -s /bin/false -d /opt/sensor-metrics-analyzer sensor-metrics
sudo chown -R sensor-metrics:sensor-metrics /opt/sensor-metrics-analyzer
```

## Step 3: Configure Systemd Service

1. Copy the service file:
   ```bash
   sudo cp /opt/sensor-metrics-analyzer/sensor-metrics-analyzer-go/web/sensor-metrics-web.service \
     /etc/systemd/system/
   ```

2. Edit the service file to match your paths:
   ```bash
   sudo nano /etc/systemd/system/sensor-metrics-web.service
   ```

   Update these paths if different:
   - `WorkingDirectory`: `/opt/sensor-metrics-analyzer/sensor-metrics-analyzer-go/web/server`
   - `ExecStart`: `/opt/sensor-metrics-analyzer/sensor-metrics-analyzer-go/bin/web-server`
   - `RULES_DIR`: `/opt/sensor-metrics-analyzer/sensor-metrics-analyzer-go/automated-rules`
   - `LOAD_LEVEL_DIR`: `/opt/sensor-metrics-analyzer/sensor-metrics-analyzer-go/automated-rules/load-level`
   - `TEMPLATE_PATH`: `/opt/sensor-metrics-analyzer/sensor-metrics-analyzer-go/templates/markdown.tmpl`

3. Reload systemd and start the service:
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable sensor-metrics-web
   sudo systemctl start sensor-metrics-web
   ```

4. Verify the service is running:
   ```bash
   sudo systemctl status sensor-metrics-web
   ```

5. Check logs:
   ```bash
   sudo journalctl -u sensor-metrics-web -f
   ```

## Step 4: Configure Nginx

1. Copy the nginx configuration:
   ```bash
   sudo cp /opt/sensor-metrics-analyzer/sensor-metrics-analyzer-go/web/nginx.conf \
     /etc/nginx/sites-available/sensor-metrics-web
   ```

2. Edit the configuration:
   ```bash
   sudo nano /etc/nginx/sites-available/sensor-metrics-web
   ```

   Update:
   - `root`: `/opt/sensor-metrics-analyzer/sensor-metrics-analyzer-go/web/static`
   - `server_name`: Your domain name (or leave as `_` for default)

3. Enable the site and disable the default nginx site:
   ```bash
   sudo rm /etc/nginx/sites-enabled/default
   sudo ln -s /etc/nginx/sites-available/sensor-metrics-web \
     /etc/nginx/sites-enabled/
   ```

4. Test nginx configuration:
   ```bash
   sudo nginx -t
   ```

5. Reload nginx:
   ```bash
   sudo systemctl reload nginx
   ```

## Step 5: Verify Deployment

1. Test the backend health endpoint directly:
   ```bash
   curl http://localhost:8080/health
   ```

2. Test the nginx health endpoint:
   ```bash
   curl http://localhost/health
   # Or: curl http://your-domain/health
   ```

3. Test the API directly:
   ```bash
   curl -X POST http://localhost/api/analyze/both \
     -F "file=@/path/to/test-metrics.prom"
   ```

4. Access the web interface:
   Open `http://your-server-ip` or `http://your-domain` in a browser

   Note: The backend on `:8080` only serves API endpoints, so `/` will return 404.

## Step 6: Firewall Configuration

If using a firewall, allow HTTP traffic:

```bash
# For UFW
sudo ufw allow 80/tcp

# For firewalld
sudo firewall-cmd --permanent --add-service=http
sudo firewall-cmd --reload
```

## Troubleshooting

### Service won't start

1. Check service status:
   ```bash
   sudo systemctl status sensor-metrics-web
   ```

2. Check logs:
   ```bash
   sudo journalctl -u sensor-metrics-web -n 50
   ```

3. Verify binary exists and is executable:
   ```bash
   ls -l /opt/sensor-metrics-analyzer/sensor-metrics-analyzer-go/bin/web-server
   ```

4. Test running manually:
   ```bash
   sudo -u sensor-metrics /opt/sensor-metrics-analyzer/sensor-metrics-analyzer-go/bin/web-server \
     --rules /opt/sensor-metrics-analyzer/sensor-metrics-analyzer-go/automated-rules
   ```

### Nginx 502 Bad Gateway

1. Verify backend is running:
   ```bash
   curl http://localhost:8080/health
   ```

2. Check nginx error logs:
   ```bash
   sudo tail -f /var/log/nginx/error.log
   ```

3. Verify proxy_pass URL in nginx config matches backend listen address

### File upload fails

1. Check nginx `client_max_body_size` setting
2. Verify file size is within limits (default: 50MB)
3. Check backend logs for detailed errors

## Security Hardening

1. **Use HTTPS**: Set up SSL/TLS certificates (Let's Encrypt recommended)
2. **Restrict access**: Use firewall rules to limit access to specific IPs if needed
3. **Regular updates**: Keep the application and system updated
4. **Monitor logs**: Set up log monitoring for suspicious activity

## Next Steps

- See [UPDATE.md](./UPDATE.md) for update procedures
- Configure log rotation if needed
- Set up monitoring/alerting for the service
