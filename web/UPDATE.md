# Update Guide

This guide covers updating the Sensor Metrics Analyzer Web service.

## Update Procedure

### Step 1: Stop the Service

```bash
sudo systemctl stop sensor-metrics-web
```

### Step 2: Backup Current Version (Optional but Recommended)

```bash
# Backup binaries
sudo cp /opt/sensor-metrics-analyzer/sensor-metrics-analyzer-go/bin/web-server \
  /opt/sensor-metrics-analyzer/sensor-metrics-analyzer-go/bin/web-server.backup
```

### Step 3: Update the Application

**Option A: Git Pull (if using git)**

```bash
cd /opt/sensor-metrics-analyzer/sensor-metrics-analyzer-go
git pull origin main  # or your branch name
```

**Option B: Manual Copy**

Copy the new files to the server, preserving the directory structure.

### Step 4: Download Updated Binary

```bash
cd /opt/sensor-metrics-analyzer/sensor-metrics-analyzer-go
curl -L -o bin/web-server https://github.com/stackrox/sensor-metrics-analyzer/releases/latest/download/web-server-linux-amd64
chmod +x bin/web-server
```

### Step 5: Update Configuration Files (if needed)

Check if any configuration files have changed:

```bash
# Compare service file
diff /opt/sensor-metrics-analyzer/sensor-metrics-analyzer-go/web/sensor-metrics-web.service \
  /etc/systemd/system/sensor-metrics-web.service

# Compare nginx config
diff /opt/sensor-metrics-analyzer/sensor-metrics-analyzer-go/web/nginx.conf \
  /etc/nginx/sites-available/sensor-metrics-web
```

If there are differences, update the files in `/etc/`:

```bash
# Update systemd service (review changes first!)
sudo cp /opt/sensor-metrics-analyzer/sensor-metrics-analyzer-go/web/sensor-metrics-web.service \
  /etc/systemd/system/
sudo systemctl daemon-reload

# Update nginx config (review changes first!)
sudo cp /opt/sensor-metrics-analyzer/sensor-metrics-analyzer-go/web/nginx.conf \
  /etc/nginx/sites-available/sensor-metrics-web
sudo nginx -t  # Test configuration
sudo systemctl reload nginx
```

### Step 6: Update Frontend Files

```bash
sudo cp -r /opt/sensor-metrics-analyzer/sensor-metrics-analyzer-go/web/static/* \
  /opt/sensor-metrics-analyzer/sensor-metrics-analyzer-go/web/static/
```

### Step 7: Verify Binaries

```bash
# Check binaries exist and are executable
ls -lh /opt/sensor-metrics-analyzer/sensor-metrics-analyzer-go/bin/

# Test web server binary (should show usage)
/opt/sensor-metrics-analyzer/sensor-metrics-analyzer-go/bin/web-server --help
```

### Step 8: Start the Service

```bash
sudo systemctl start sensor-metrics-web
sudo systemctl status sensor-metrics-web
```

### Step 9: Verify Health

```bash
# Check health endpoint
curl http://localhost:8080/health

# Check service logs
sudo journalctl -u sensor-metrics-web -f
```

### Step 10: Test the Web Interface

1. Open the web interface in a browser
2. Upload a test metrics file
3. Verify both console and markdown outputs are generated correctly

## Rollback Procedure

If something goes wrong, rollback to the previous version:

```bash
# Stop service
sudo systemctl stop sensor-metrics-web

# Restore binaries
sudo cp /opt/sensor-metrics-analyzer/sensor-metrics-analyzer-go/bin/web-server.backup \
  /opt/sensor-metrics-analyzer/sensor-metrics-analyzer-go/bin/web-server

# Start service
sudo systemctl start sensor-metrics-web
```

## Automated Update Script

You can create a simple update script:

```bash
#!/bin/bash
# /opt/sensor-metrics-analyzer/update.sh

set -e

SERVICE_NAME="sensor-metrics-web"
APP_DIR="/opt/sensor-metrics-analyzer/sensor-metrics-analyzer-go"
BIN_DIR="$APP_DIR/bin"

echo "Stopping service..."
sudo systemctl stop $SERVICE_NAME

echo "Backing up binaries..."
sudo cp $BIN_DIR/web-server $BIN_DIR/web-server.backup.$(date +%Y%m%d_%H%M%S)

echo "Updating application..."
cd $APP_DIR
# git pull  # Uncomment if using git

echo "Downloading binary..."
curl -L -o $BIN_DIR/web-server https://github.com/stackrox/sensor-metrics-analyzer/releases/latest/download/web-server-linux-amd64
chmod +x $BIN_DIR/web-server

echo "Starting service..."
sudo systemctl start $SERVICE_NAME

echo "Waiting for service to start..."
sleep 2

echo "Checking service status..."
sudo systemctl status $SERVICE_NAME --no-pager

echo "Testing health endpoint..."
curl -s http://localhost:8080/health || echo "Health check failed!"

echo "Update complete!"
```

Make it executable:
```bash
chmod +x /opt/sensor-metrics-analyzer/update.sh
```

## Update Checklist

- [ ] Stop the service
- [ ] Backup current binaries
- [ ] Update application files
- [ ] Rebuild binaries
- [ ] Update configuration files (if changed)
- [ ] Update frontend files
- [ ] Verify binaries
- [ ] Start the service
- [ ] Verify health endpoint
- [ ] Test web interface
- [ ] Monitor logs for errors

## Troubleshooting Updates

### Service fails to start after update

1. Check logs: `sudo journalctl -u sensor-metrics-web -n 50`
2. Verify binary exists: `ls -l /opt/sensor-metrics-analyzer/sensor-metrics-analyzer-go/bin/web-server`
3. Test binary manually: `sudo -u sensor-metrics /opt/sensor-metrics-analyzer/sensor-metrics-analyzer-go/bin/web-server --help`
4. Rollback if needed

### Binary not found errors

- Verify the build completed successfully
- Check file permissions: `sudo chown sensor-metrics:sensor-metrics /opt/sensor-metrics-analyzer/sensor-metrics-analyzer-go/bin/*`
- Ensure binaries are executable: `sudo chmod +x /opt/sensor-metrics-analyzer/sensor-metrics-analyzer-go/bin/*`

### Configuration errors

- Review configuration file changes before applying
- Test nginx config: `sudo nginx -t`
- Reload systemd after service file changes: `sudo systemctl daemon-reload`
