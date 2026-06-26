# HOWTO — Zabbix 7 Setup

Step-by-step guide to monitor domain expiration with the `check_domain` binary on **Zabbix 7.x**.

> The `zbx_domain_expiry.yaml` template was exported in Zabbix 6.4 format and is compatible with Zabbix 7. **External check** items are executed by the **Zabbix Server** (or by a **Proxy**, if the monitored host is monitored through a proxy).

## Table of contents

1. [Prerequisites](#1-prerequisites)
2. [Install the binary](#2-install-the-binary)
3. [Configure the Zabbix Server](#3-configure-the-zabbix-server)
4. [Import the template](#4-import-the-template)
5. [Create the domain host](#5-create-the-domain-host)
6. [Test data collection](#6-test-data-collection)
7. [Adjust macros and alerts](#7-adjust-macros-and-alerts)
8. [Monitor multiple domains](#8-monitor-multiple-domains)
9. [Troubleshooting](#9-troubleshooting)

---

## 1. Prerequisites

| Item | Requirement |
|------|-------------|
| Zabbix | 7.0 or higher |
| Zabbix server OS | GNU/Linux **amd64** |
| Access | Shell on the Zabbix server (or on the proxy responsible for the host) |
| Network | Outbound HTTPS (RDAP) and TCP/43 (WHOIS) allowed |

No need to install `curl`, `jq`, `whois`, or other tools — the Go binary is self-contained.

---

## 2. Install the binary

### Option A — Release (recommended)

On the **Zabbix server** (or proxy, if applicable):

```bash
VERSION=0.0.1
curl -LO "https://github.com/jniltinho/zabbix-domain-expiry/releases/download/v${VERSION}/check_domain_${VERSION}_linux_amd64.tar.gz"
tar -xzf "check_domain_${VERSION}_linux_amd64.tar.gz"
sudo install -m 755 -o zabbix -g zabbix check_domain /usr/lib/zabbix/externalscripts/check_domain
```

### Option B — Build on the server

```bash
git clone https://github.com/jniltinho/zabbix-domain-expiry.git
cd zabbix-domain-expiry
make build-linux-amd64
sudo install -m 755 -o zabbix -g zabbix build/check_domain-linux-amd64 /usr/lib/zabbix/externalscripts/check_domain
```

### Validate manually

Run as the `zabbix` user to confirm permissions and connectivity:

```bash
sudo -u zabbix /usr/lib/zabbix/externalscripts/check_domain -d example.com -r 0 -s 0 -w 30 -c 7
```

Expected output (JSON on stdout):

```json
{"state":"OK","days_left":365,"days_since_expired":0,"expire_date":"2026-08-13","message":"State: OK ; Days left: 365 ; Expire date: 2026-08-13"}
```

---

## 3. Configure the Zabbix Server

### External scripts directory

Confirm the path in `/etc/zabbix/zabbix_server.conf`:

```ini
### Option: ExternalScripts
ExternalScripts=/usr/lib/zabbix/externalscripts
```

If you change the path, restart the server:

```bash
sudo systemctl restart zabbix-server
```

### Zabbix Proxy

If the domain host is monitored by a **proxy**, the binary must be installed on the **proxy**, not only on the central server. Set `ExternalScripts` in `/etc/zabbix/zabbix_proxy.conf` and restart the proxy:

```bash
sudo systemctl restart zabbix-proxy
```

### Permissions

Zabbix runs external scripts as the `zabbix` user. Ensure:

```bash
ls -l /usr/lib/zabbix/externalscripts/check_domain
# -rwxr-xr-x 1 zabbix zabbix ... check_domain
```

---

## 4. Import the template

1. Open the Zabbix 7 web interface
2. Go to **Data collection → Templates**
3. Click **Import** (top right)
4. Select `zbx_domain_expiry.yaml`
5. On the import screen, enable:
   - **Templates** — Create new / Update existing (as needed)
   - **Items**, **Triggers**, **Template groups**
6. Click **Import**

After import, the **Domain Expiry** template should appear under **Data collection → Templates**.

### Upgrading an existing template

When upgrading from older versions (shell script `check_domain.sh`):

1. Replace the binary in `externalscripts/`
2. Re-import `zbx_domain_expiry.yaml` with **Update existing** enabled
3. Confirm the master item key changed from `check_domain.sh[...]` to `check_domain[...]`

---

## 5. Create the domain host

The template uses `{HOST.HOST}` as the domain name. The Zabbix **Host name** must be the exact domain to monitor.

1. Go to **Data collection → Hosts**
2. Click **Create host**
3. Fill in:

| Field | Example | Notes |
|-------|---------|-------|
| **Host name** | `example.com` | Must be the actual domain |
| **Visible name** | `Example.com expiry` | Display only |
| **Host groups** | `Domains` (or other) | Optional |
| **Interfaces** | — | No interface required for external checks |

4. On the **Templates** tab, add the **Domain Expiry** template
5. Click **Add** / **Update**

> **Important:** do not add prefixes to the Host name (e.g. `zabbix.example.com` instead of `example.com`) unless that is the domain you intend to query.

---

## 6. Test data collection

### Execute the item manually

1. Open the created host → **Items** tab
2. Find the **Check Domain** item (External check type)
3. Select the item and click **Execute now**
4. Wait a few seconds and check **Latest data**

### Expected items

| Item | Expected value |
|------|----------------|
| Check Domain | Full JSON response |
| State | `OK`, `WARNING`, `CRITICAL`, or `UNKNOWN` |
| Days Left | Number of days remaining |
| Expire Date | Date in `YYYY-MM-DD` format |
| Message | Descriptive status message |

### Verify triggers

Under **Monitoring → Problems**, confirm there are no false alerts after the first successful collection.

---

## 7. Adjust macros and alerts

Macros can be set on the template or overridden per host.

Go to **Data collection → Templates → Domain Expiry → Macros** (or host-level macros).

| Macro | Default | Description |
|-------|---------|-------------|
| `{$EXP_CRIT}` | `7` | Days remaining before a **High** alert |
| `{$EXP_WARN}` | `30` | Days remaining before a **Warning** alert |
| `{$RDAP_SERVER}` | *(empty)* | RDAP server URL; empty uses IANA bootstrap |
| `{$WHOIS_SERVER}` | *(empty)* | WHOIS server; empty uses built-in lookup |

### Example — custom RDAP for a TLD

For a `.uk` domain, if needed:

```
{$RDAP_SERVER} = https://rdap.nominet.uk/uk-domain/
```

Leave empty for most domains — the binary resolves the RDAP server automatically.

### Included triggers

| Priority | Condition |
|----------|-----------|
| Not classified | State is `UNKNOWN` (lookup failure) |
| Disaster | Domain has expired |
| High | Expires in ≤ `{$EXP_CRIT}` days |
| Warning | Expires in ≤ `{$EXP_WARN}` days |

---

## 8. Monitor multiple domains

One domain = **one host** in Zabbix.

```
Host: example.com     → Domain Expiry template
Host: meusite.com.br  → Domain Expiry template
Host: outro.org       → Domain Expiry template
```

The default collection interval is **1 day** (`1d`), which helps avoid WHOIS/RDAP rate limits. To change it:

1. Open the **Check Domain** item on the template or host
2. Adjust **Update interval** (e.g. `12h`, `1d`)

---

## 9. Troubleshooting

### Item in NOT SUPPORTED state

| Cause | Solution |
|-------|----------|
| Binary missing | Install at `/usr/lib/zabbix/externalscripts/check_domain` |
| No execute permission | `chmod 755` and owner `zabbix:zabbix` |
| Wrong `ExternalScripts` path | Check `zabbix_server.conf` / `zabbix_proxy.conf` |
| Script on wrong server | Install on the proxy if the host uses a proxy |

### State = UNKNOWN

Test manually with debug:

```bash
sudo -u zabbix /usr/lib/zabbix/externalscripts/check_domain -d example.com -r 0 -s 0 -z
```

Debug messages go to **stderr**; the JSON result is still written to stdout.

Common causes:

- Domain does not exist or has no public expiration data
- Network blocked (firewall without outbound HTTPS/43)
- WHOIS server rate limit (wait or increase collection interval)
- TLD with non-standard RDAP URL format

### JSONPath does not extract values

Confirm the **Check Domain** item returns valid JSON in **Latest data**. Dependent items (`Days Left`, `State`, etc.) rely on this master item.

### Zabbix logs

```bash
# RHEL/Rocky
sudo tail -f /var/log/zabbix/zabbix_server.log

# Debian/Ubuntu
sudo tail -f /var/log/zabbix-server/zabbix_server.log
```

Look for errors related to `check_domain` or `External script`.

### Test RDAP/WHOIS connectivity

```bash
curl -s "https://rdap.org/domain/example.com" | head
whois example.com | head
```

---

## References

- [Zabbix 7 — External check](https://www.zabbix.com/documentation/7.0/en/manual/config/items/itemtypes/external)
- [Zabbix 7 — Template import](https://www.zabbix.com/documentation/7.0/en/manual/xml_export_import/templates)
- [Project repository](https://github.com/jniltinho/zabbix-domain-expiry)
- [README](README.md)