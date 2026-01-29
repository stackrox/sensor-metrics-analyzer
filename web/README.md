# Sensor Metrics Analyzer Web

Web interface for the Sensor Metrics Analyzer, allowing users to upload Prometheus metrics files and view analysis reports in both console and markdown formats.

## Architecture

- **Backend**: Go HTTP server running as a systemd service
- **Frontend**: Static HTML/JavaScript served by Nginx
- **API**: RESTful endpoint `/api/analyze/both` that accepts file uploads and returns JSON with both console and markdown outputs

## Components

- `server/` - Go HTTP server source code
- `static/` - Frontend HTML/JS/CSS files
- `nginx.conf` - Nginx configuration template
- `sensor-metrics-web.service` - Systemd service unit file

## Quick Start

### Prerequisites

- Nginx installed and running
- Linux system with systemd

### Download Precompiled Binaries

We publish precompiled binaries in GitHub Releases. Download the latest `web-server` binary for your platform and place it in `sensor-metrics-analyzer-go/bin/`.

Example:
```bash
curl -L -o bin/web-server https://github.com/stackrox/sensor-metrics-analyzer/releases/latest/download/web-server-linux-amd64
chmod +x bin/web-server
```

### Deployment

See [DEPLOYMENT.md](./DEPLOYMENT.md) for detailed deployment instructions.

### Updating

See [UPDATE.md](./UPDATE.md) for update procedures.

## Development

### Running Locally

1. Start the backend server:
   ```bash
   cd sensor-metrics-analyzer-go/web/server
   go run main.go \
     --listen :8080 \
     --rules ../automated-rules \
     --load-level-dir ../automated-rules/load-level
   ```

2. Serve the frontend (in another terminal):
   ```bash
   cd sensor-metrics-analyzer-go/web/static
   python3 -m http.server 8000
   ```

3. Access the web interface at `http://localhost:8000`

   Note: For local development, you may need to update the API endpoint in `index.html` to point to `http://localhost:8080/api/analyze/both` or use a local proxy.

### Testing

Test the API endpoint directly:
```bash
curl -X POST http://localhost:8080/api/analyze/both \
  -F "file=@/path/to/metrics.prom"
```

## Configuration

The web server can be configured via command-line flags or environment variables:

- `--listen` / `LISTEN_ADDR`: Listen address (default: `:8080`)
- `--rules` / `RULES_DIR`: Rules directory (default: `./automated-rules`)
- `--load-level-dir` / `LOAD_LEVEL_DIR`: Load level rules directory (default: `./automated-rules/load-level`)
- `--template` / `TEMPLATE_PATH`: Path to markdown template (default: `./templates/markdown.tmpl`)
- `--max-size` / `MAX_FILE_SIZE`: Maximum upload size in bytes (default: 50MB)
- `--timeout` / `REQUEST_TIMEOUT`: Request timeout duration (default: 60s)

## API Endpoints

### POST /api/analyze/both

Upload a metrics file and receive both console and markdown outputs.

**Request:**
- Method: `POST`
- Content-Type: `multipart/form-data`
- Body: Form field `file` containing the metrics file

**Response:**
```json
{
  "console": "...",
  "markdown": "...",
  "error": ""  // Optional, present if there were errors
}
```

### GET /health

Health check endpoint.

**Response:**
```json
{
  "status": "ok"
}
```

## Troubleshooting

### Server won't start

- Verify the rules directory exists and contains valid TOML files
- Check systemd logs: `journalctl -u sensor-metrics-web -f`

### File upload fails

- Check nginx `client_max_body_size` setting
- Verify the file size is within limits
- Check server logs for detailed error messages

### Analysis returns errors

- Ensure the uploaded file is a valid Prometheus metrics file
- Check that rules are properly configured
- Review server logs for analyzer output

## Security Considerations

- The service runs as a dedicated user (`sensor-metrics`)
- Temporary files are automatically cleaned up after processing
- File size limits prevent resource exhaustion
- Request timeouts prevent long-running requests
- No persistent storage of uploaded files or reports
