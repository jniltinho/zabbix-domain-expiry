# zabbix-domain-expiry

Monitor domain expiration dates in Zabbix using RDAP or WHOIS.

Version **2.0.0** ships a self-contained Go binary (`check_domain`) that replaces the original shell script. No runtime dependencies — networking, RDAP/WHOIS parsing, and JSON output are all built in using the Go standard library.

## Features

- **RDAP and WHOIS** — queries expiration via RDAP (preferred) with automatic fallback to WHOIS
- **Zero runtime dependencies** — static binary, no `curl`, `jq`, `whois`, or other tools required
- **JSON output** — structured response for Zabbix JSONPath preprocessing
- **Debug mode** — detailed diagnostics written to stderr
- **Zabbix template** — ready-to-import template with items, triggers, and macros

## Requirements

| Component | Version |
|-----------|---------|
| Zabbix Server/Agent | 6.4 or higher |
| OS | GNU/Linux (amd64 or arm64) |

To **build from source**, Go 1.21 or later is required.

## Quick Start

```bash
git clone https://github.com/a-stoyanov/zabbix-domain-expiry.git
cd zabbix-domain-expiry
make install          # builds and installs to /usr/lib/zabbix/externalscripts/
```

Import `zbx_domain_expiry.yaml` into Zabbix, then create a host named after the domain (e.g. `example.com`) and attach the **Domain Expiry** template.

## Installation

### Option 1 — Build and install (recommended)

```bash
make install
```

This compiles a static binary and copies it to `/usr/lib/zabbix/externalscripts/check_domain`.

### Option 2 — Cross-compile

Build for a specific platform without UPX compression:

```bash
make build-linux-amd64    # build/check_domain-linux-amd64
make build-linux-arm64    # build/check_domain-linux-arm64
make build-all            # both targets
```

Copy the appropriate binary to the Zabbix external scripts directory:

```bash
install -m 755 build/check_domain-linux-amd64 /usr/lib/zabbix/externalscripts/check_domain
```

### Option 3 — Download a release binary

If a pre-built binary is available from the [releases page](https://github.com/a-stoyanov/zabbix-domain-expiry/releases), copy it directly:

```bash
install -m 755 check_domain /usr/lib/zabbix/externalscripts/check_domain
```

### Zabbix template

1. In Zabbix, go to **Data collection → Templates → Import** and upload `zbx_domain_expiry.yaml`
2. Create a host with the domain as the host name (e.g. `example.com`)
3. Link the **Domain Expiry** template to the host

## Upgrading from v1.x (shell script)

1. Import/overwrite the template (`zbx_domain_expiry.yaml`) — the external check item key changed from `check_domain.sh[...]` to `check_domain[...]`
2. Replace `check_domain.sh` with the new `check_domain` binary in `/usr/lib/zabbix/externalscripts/`
3. Remove shell script dependencies (`curl`, `jq`, `whois`, etc.) — they are no longer needed

The legacy shell script (`check_domain.sh`) is kept in the repository for reference but is no longer maintained.

## Usage

```bash
check_domain -d example.com
check_domain -d example.com -r 'https://rdap.example.com' -s whois.example.com -w 30 -c 7
check_domain -d example.com -z          # debug output to stderr
check_domain -V                       # print version as JSON
```

### Options

| Flag | Description |
|------|-------------|
| `-d`, `--domain` | Domain name to check (**required**) |
| `-w`, `--warning` | Warning threshold in days (default: `30`) |
| `-c`, `--critical` | Critical threshold in days (default: `7`) |
| `-r`, `--rdap-server` | RDAP server URL (`""` or `0` for IANA bootstrap lookup) |
| `-s`, `--whois-server` | WHOIS server hostname (`""` or `0` for default lookup) |
| `-P`, `--path` | Accepted for backward compatibility; ignored by the Go binary |
| `-z`, `--debug` | Enable debug output to stderr |
| `-h`, `--help` | Display help |
| `-V`, `--version` | Display version in JSON format |

### Example output

```json
{"state":"OK","days_left":365,"days_since_expired":0,"expire_date":"2026-06-24","message":"State: OK ; Days left: 365 ; Expire date: 2026-06-24"}
```

### Exit codes

| Code | State | Meaning |
|------|-------|---------|
| `0` | OK | Domain is valid and not near expiration |
| `1` | WARNING | Days remaining ≤ warning threshold |
| `2` | CRITICAL | Days remaining ≤ critical threshold, or domain has expired |
| `3` | UNKNOWN | Lookup or parsing failed |

## Zabbix Template Reference

### Macros

| Macro | Default | Description |
|-------|---------|-------------|
| `{$EXP_CRIT}` | `7` | Days remaining before a **High** alert |
| `{$EXP_WARN}` | `30` | Days remaining before a **Warning** alert |
| `{$RDAP_SERVER}` | *(empty)* | Override RDAP server; empty uses IANA bootstrap |
| `{$WHOIS_SERVER}` | *(empty)* | Override WHOIS server; empty uses built-in lookup |

### Items

| Name | Key | Type | Description |
|------|-----|------|-------------|
| Check Domain | `check_domain[...]` | External | Runs the external check script |
| Days Left | `check_domain.days_left` | Dependent | Days until expiration |
| Days Since Expired | `check_domain.days_since_expired` | Dependent | Days since expiration (0 if active) |
| Expire Date | `check_domain.expire_date` | Dependent | Expiration date (`YYYY-MM-DD`) |
| Message | `check_domain.message` | Dependent | Status message from the check |
| State | `check_domain.state` | Dependent | `OK`, `WARNING`, `CRITICAL`, or `UNKNOWN` |

### Triggers

| Priority | Condition |
|----------|-----------|
| Not classified | State is `UNKNOWN` |
| Disaster | Domain has expired |
| High | State is `CRITICAL` and days left ≤ `{$EXP_CRIT}` |
| Warning | State is `WARNING` and days left ≤ `{$EXP_WARN}` |

## Development

### Project layout

```
.
├── main.go                  # entry point
├── internal/
│   ├── checkdomain/         # CLI parsing and check orchestration
│   ├── output/              # JSON output and exit codes
│   ├── rdap/                # RDAP client (IANA bootstrap + queries)
│   └── whois/               # WHOIS client and date parsing
├── zbx_domain_expiry.yaml   # Zabbix template
├── Makefile
└── check_domain.sh          # legacy shell script (deprecated)
```

### Makefile targets

```bash
make build              # static binary with UPX compression (if available)
make build-nocompress   # static binary without UPX
make build-all          # cross-compile linux/amd64 and linux/arm64
make test               # run unit tests
make run ARGS='-d example.com'   # build and run
make clean              # remove build artifacts
```

The binary is built with `CGO_ENABLED=0` for a fully static, portable executable.

## Debugging

```bash
check_domain -d example.com -z
```

Debug messages are written to stderr; stdout always contains the JSON result. Check Zabbix server logs if the external check fails silently.

Verify RDAP/WHOIS responses independently when troubleshooting UNKNOWN states:

```bash
curl -s "https://rdap.org/domain/example.com" | jq .
whois example.com
```

## Notes

- RDAP is tried first for faster, structured queries; WHOIS is used as fallback
- WHOIS date parsing supports common registrar formats but may fail on non-standard responses
- WHOIS rate limits can cause `UNKNOWN` states — the default 1-day check interval is conservative
- Some TLDs (e.g. `.uk`, `.br`) use non-standard RDAP URL paths; see `adjustURL()` in `internal/rdap/rdap.go`

## License

This project is licensed under the [Apache License 2.0](LICENSE).