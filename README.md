# logslice

Lightweight log aggregation proxy that filters, tags, and forwards structured logs to multiple sinks simultaneously.

---

## Installation

```bash
go install github.com/yourorg/logslice@latest
```

Or build from source:

```bash
git clone https://github.com/yourorg/logslice.git && cd logslice && go build ./...
```

---

## Usage

Start the proxy with a config file:

```bash
logslice --config logslice.yaml
```

Example `logslice.yaml`:

```yaml
listen: ":5170"

filters:
  - field: level
    match: "error|warn"

tags:
  env: production
  service: api-gateway

sinks:
  - type: stdout
  - type: file
    path: /var/log/app/errors.log
  - type: http
    url: https://logs.example.com/ingest
```

Send a structured log entry:

```bash
echo '{"level":"error","msg":"something failed","code":500}' | nc localhost 5170
```

logslice will filter the entry, attach the configured tags, and forward it to all defined sinks concurrently.

---

## Configuration Reference

| Key | Description |
|-----|-------------|
| `listen` | Address and port the proxy listens on |
| `filters` | Field-based rules to include or drop log entries |
| `tags` | Key-value pairs appended to every forwarded entry |
| `sinks` | List of output destinations (`stdout`, `file`, `http`) |

---

## License

MIT © yourorg