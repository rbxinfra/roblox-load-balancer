# Roblox Load Balancer (RLB)

A Go daemon that dynamically builds and manages an HAProxy configuration from services registered in [Consul](https://www.consul.io/). It watches the Consul catalog, turns per-service tags into HAProxy frontend/backend rules, renders them into a user-supplied HAProxy template, and keeps the running HAProxy process in sync — reloading it via `SIGHUP` (or starting it if it isn't running yet) whenever the set of backends changes.

It is conceptually similar to tools like Traefik or Consul Template, but purpose-built to drive HAProxy from Consul service tags with minimal moving parts.

> [!NOTE]
> **This project is not affiliated with Roblox Corporation.** See the [Notice](#notice) section at the bottom of this document.

## How it works

On each refresh cycle (or when triggered manually), the daemon:

1. **Queries Consul** for every service instance tagged `<prefix>.enable=true` (`prefix` defaults to `haproxy`, see [Configuration](#configuration-file)).
2. **Parses each service's tags** into a strongly typed `ServiceConfig` (protocol, frontend rules, backend rules — see [Service Tags](#service-tags-consul-labels)), applying defaults and validating them (e.g. FQDN is required, entrypoints must exist).
3. **Builds HAProxy `backend` blocks and frontend routing rules** for every configured entrypoint, based on the discovered services and their nodes.
4. **Renders these into your HAProxy template** using two template functions, `backends "<entrypoint>"` and `rules "<entrypoint>"`, and writes the result to the configured output file.
5. **Diffs the new set of services against the last known set** (via a cheap hash) and, if anything changed, validates the new HAProxy config (`haproxy -c`) and reloads HAProxy:
   - If HAProxy is already running, it sends `SIGHUP` (graceful reload, no dropped connections).
   - If it isn't running, it starts it (retrying up to `HAProxy.MaxStartAttempts` times).
6. **Sleeps until `RefreshInterval` elapses** or a refresh is triggered early, then repeats.

A refresh can also be triggered outside the normal interval by:
- Sending the daemon process `SIGUSR1`.
- A change being detected on any file listed in `ReloadOnChangesDetectedForFiles` (e.g. a TLS certificate being renewed) — the daemon polls for the file/directory to exist and then watches it with `fsnotify`.

On shutdown (`SIGABRT`/`SIGINT`/`SIGTERM`), the daemon signals its background goroutines to stop and sends `SIGUSR1` to HAProxy for a graceful shutdown.

## Building

Ensure you have [Go 1.25+](https://go.dev/dl/) and [GNU Make](https://www.gnu.org/software/make/).

1. Clone the repository via `git`:

    ```txt
    git clone git@github.rbx.com:Roblox/roblox-load-balancer.git
    cd roblox-load-balancer
    ```

2. Build via `make`:

    ```txt
    make build-debug
    ```

    Other useful targets:

    | Target | Description |
    | --- | --- |
    | `make build-debug` | Debug build for the current OS/arch. |
    | `make build-release` | Release build (symbols stripped) for the current OS/arch. |
    | `make build-debug-<arch>` / `make build-release-<arch>` | Cross-build for a specific architecture (`x86`, `x64`, `arm`, `arm64`); `GOOS` follows the host unless overridden. |
    | `make build-debug-all` / `make build-release-all` | Build for every supported architecture. |
    | `make test` | Runs `go fmt`, `go vet`, and `go test` against `./src`. |
    | `make vendor` | Runs `go mod tidy && go mod vendor` (skip with `SKIP_VENDOR=1`). |

   Binaries are written to `bin/<debug|release>/<GOOS>/<arch>/`.

### Docker

The provided `Dockerfile` builds RLB from source on top of `haproxy:3.3-alpine` and runs it as the container's HAProxy process:

```txt
docker build -t roblox-load-balancer .
```

The container entrypoint is the compiled `roblox-load-balancer-daemon` binary itself — RLB starts, manages, and reloads the `haproxy` binary already present in the base image.

## Usage

```txt
cd src && go run main.go --help
```

(use the built binary in `bin/` instead if you downloaded or built one)

```txt
Usage: roblox-load-balancer
Build Mode: debug
Commit:
        [-h|--help]
        [--configuration-file-path[=]] [--dry-run]

  -alsologtostderr
        log to standard error as well as files
  -configuration-file-path string
        The path to the static configuration.
  -dry-run
        Reads from Consul, builds the config, and outputs to the file without starting the Daemon or reloading HAProxy.
  -help
        Print usage.
  -log_backtrace_at value
        when logging hits line file:N, emit a stack trace
  -log_dir string
        If non-empty, write log files in this directory
  -log_link string
        If non-empty, add symbolic links in this directory to the log files
  -logbuflevel int
        Buffer log messages logged at this level or lower (-1 means don't buffer; 0 means buffer INFO only; ...). Has limited applicability on non-prod platforms.
  -logtostderr
        log to standard error instead of files
  -stderrthreshold value
        logs at or above this threshold go to stderr (default 2)
  -v value
        log level for V logs
  -vmodule value
        comma-separated list of pattern=N settings for file-filtered logging
```

`--configuration-file-path` is required and must point to a `.json`, `.yml`/`.yaml`, or `.toml` file (see below). `--dry-run` fetches from Consul, builds the config once, and writes it to `OutputFilePath` without starting the daemon loop or touching a running HAProxy process — useful for previewing what would be generated.

## Configuration file

The daemon needs a static configuration file (JSON, YAML, or TOML — selected by file extension) describing where to find its HAProxy template, Consul connection details, and defaults applied to every discovered service. Top-level keys (YAML names shown; JSON/TOML use the same structure with `camelCase`/`snake_case` respectively):

| Key | Type | Default | Description |
| --- | --- | --- | --- |
| `prefix` | string | `haproxy` | Tag prefix used both to filter Consul services (`<prefix>.enable=true`) and to parse per-service tags. |
| `template_file_path` | string | *required* | Path to the HAProxy config **template** (a Go `text/template` file). Relative paths are resolved to absolute at load time, and the file must exist. |
| `output_file_path` | string | `/usr/local/etc/haproxy/haproxy.cfg` | Where the rendered HAProxy config is written. |
| `reload_on_changes_detected_for_files` | []string | — | Extra files to watch (e.g. a TLS bundle); any write/removal triggers a manual refresh + reload. |
| `refresh_interval` | duration | `5m` | How often to poll Consul and re-render, absent an earlier manual trigger. |
| `tls_bundle_file_path` | string | — | CA bundle used to verify TLS backends where a service's protocol is `https` (see [Service Tags](#service-tags-consul-labels)). |
| `entrypoints` | map[string]EntrypointConfig | — | **Required, at least one entry.** Named entrypoints (e.g. `http`, `https`) that services can opt into; each can inject extra request headers into every backend using that entrypoint (see below). |
| `health_checks` | map[string]HealthCheckConfig | — | Per-service (keyed by Consul service name) or `default` HAProxy active health-check configuration (see below). |
| `servers` | ServersConfig | see below | Default health-check `interval`/`rise`/`fall` applied to backend servers. |
| `consul` | ConsulConfig | — | Consul API client configuration (address, scheme, ACL token, TLS, namespace/partition, etc. — a thin wrapper around `hashicorp/consul/api`'s `Config`). |
| `haproxy` | HAProxyConfig | see below | How to invoke and supervise the HAProxy process itself. |

### `entrypoints`

Each entrypoint is a named routing context (`http`, `https`, an internal-only entrypoint, etc.) that a service can be exposed on via its `fe.entrypoints` tag. An entrypoint can also define headers to inject into **every** backend using it:

```yaml
entrypoints:
  http:
    request_headers:
      X-Forwarded-Proto:
        value: "http"
        append_value: false
  https: {}
```

If a service doesn't specify `fe.entrypoints`, it is attached to **all** configured entrypoints by default.

### `servers`

Controls the default HAProxy health-check cadence:

```yaml
servers:
  default:               # applied server-wide via `default-server` (Rise/Fall combine with per_server timing)
    interval: 5s
    rise: 2
    fall: 3
  per_server:             # applied to each individual `server` line
    interval: 10s
    rise: 1
    fall: 1
```

These are the built-in defaults if `servers` is omitted entirely.

### `health_checks`

Configures active HTTP health checking (HAProxy's `option httpchk` / `http-check send` / `http-check expect`). Keys are Consul service names, with a special `default` key applied to every service that doesn't have its own entry:

```yaml
health_checks:
  default:
    option: { enabled: true, method: "GET", uri: "/health" }
```

If `enabled` is false (or the key is absent for a service), no active health check directives are emitted. If `option.method`/`option.version` are omitted they default to `HEAD` / `HTTP/1.1`. If no `expect` directives are given, `expect status 200` is assumed. If no `send` directive is given, one is generated automatically with a `Host` header set to the service's first FQDN — this ensures name-based virtual hosting still works during health checks.

### `haproxy`

Controls how the daemon runs HAProxy itself:

```yaml
haproxy:
  path: /usr/sbin/haproxy          # defaults to whatever `haproxy` resolves to on $PATH
  args: ["-W", "-db", "-f", "/usr/local/etc/haproxy/haproxy.cfg"]  # defaults to -W -db -f <output_file_path>
  stdout_log_file_path: /var/log/haproxy/stdout
  stderr_log_file_path: /var/log/haproxy/stderr
  max_start_attempts: 3
```

At startup, RLB also tries to **adopt an already-running HAProxy process** (matched by executable path) rather than starting a duplicate — useful across daemon restarts/deploys.

## Service tags (Consul labels)

Services opt into RLB by adding tags to their Consul service registration, using the configured `prefix` (default `haproxy`). Only instances tagged `<prefix>.enable=true` are considered at all; the daemon fetches them with a server-side Consul catalog filter, so untagged services never even reach the daemon.

Tags are parsed with [`traefik/paerser`](https://github.com/traefik/paerser) (the same label-parsing library Traefik uses for its own provider labels), mapping dot-delimited tag keys onto the `ServiceConfig` struct:

| Tag | Field | Notes |
| --- | --- | --- |
| `<prefix>.enable` | — | Must be `true` for the service to be picked up at all. |
| `<prefix>.protocol` | `Protocol` | One of `http` (default), `https`, `h2c`. |
| `<prefix>.fe.fqdn` | `Fe.Fqdn` | **Required.** One or more hostnames this service should be routed on (used for `hdr(host)` matching). |
| `<prefix>.fe.entrypoints` | `Fe.EntryPoints` | Which configured entrypoints this service is exposed on. Defaults to *all* entrypoints if omitted. Must reference entrypoints declared in the daemon config. |
| `<prefix>.fe.pathprefix` | `Fe.PathPrefix` | If set, only requests under this path prefix are routed to the backend, and the prefix is stripped before forwarding. |
| `<prefix>.fe.blockedpaths` | `Fe.BlockedPaths` | Exact paths that fall through to the *default* (404) backend instead of this one. |
| `<prefix>.fe.blockedpaths_beg` | `Fe.BlockedPaths_Beg` | Same as above, but matched as path prefixes. |
| `<prefix>.be.balance` | `Be.Balance` | HAProxy `balance` algorithm. Defaults to `roundrobin`. |
| `<prefix>.be.hashtype` | `Be.HashType` | HAProxy `hash-type`. Defaults to `consistent`. |
| `<prefix>.be.blockedpaths` | `Be.BlockedPaths` | Exact paths that get an HAProxy `http-request deny` (403) inside the backend. |
| `<prefix>.be.blockedpaths_beg` | `Be.BlockedPaths_Beg` | Same as above, matched as path prefixes. |
| `<prefix>.be.delheaders` | `Be.Del_Headers` | Request headers to strip before forwarding to the backend. |
| `<prefix>.be.sethostheader` | `Be.SetHostHeader` | Overrides the `Host` header sent to the backend. |

Example Consul service registration tags:

```json
{
  "Tags": [
    "haproxy.enable=true",
    "haproxy.protocol=http",
    "haproxy.fe.fqdn=api.example.com",
    "haproxy.fe.entrypoints=http,https",
    "haproxy.be.balance=leastconn"
  ]
}
```

> Field names in the table above follow the underlying Go struct; exact tag casing/formatting for list values follows `paerser`'s conventions (comma-separated lists), so it's worth validating a new tag combination with `--dry-run` before rolling it out.

### Node naming for Nomad-registered services

If a service instance's Consul metadata includes `external-source=nomad`, RLB derives its display node name from the short Nomad allocation ID embedded in the Consul service ID, instead of the raw Consul node name — making generated HAProxy `server` lines easier to correlate with Nomad allocations.

## The HAProxy template

`template_file_path` points at a Go `text/template` file that becomes the actual HAProxy config after rendering. RLB injects two functions your template can call once per entrypoint:

- `{{ backends "http" }}` — expands to every backend block (`backend <service>.<entrypoint> { ... }`) for services attached to the `http` entrypoint.
- `{{ rules "http" }}` — expands to the `use_backend ... if { hdr(host) -i ... }` routing rules (plus any path-prefix/blocked-path conditions) for the same entrypoint, intended for use inside a `frontend` block.

A minimal template:

```haproxy
frontend http
  bind *:80

  default_backend default.http
  {{ rules "http" }}

backend default.http
  http-request return errorfile /local/metadata/errorfiles/404.http

{{ backends "http" }}
```

Generated backends automatically include: `default-server pool-purge-delay 30s`, any entrypoint-level headers, path-blocking/prefix rules, header deletion/host overrides from the service's tags, `balance`/`hash-type`, health-check directives, `default-server inter <interval> rise <rise> fall <fall>`, and one `server` line per Consul node (with TLS options for `https`, `proto h2` for `h2c`).

See `rlb-public.nomad` / `rlb-private.nomad` (below) for a complete, production-grade template including stats/Prometheus frontends, error pages, and a health-check endpoint.

## Reference deployments (Nomad)

Two example [HashiCorp Nomad](https://www.nomadproject.io/) job specifications are included as a reference for running RLB in production, one per network exposure:

[See this gist](https://gist.github.com/nikita-petko/9d5344ee0e9fb4cfc88a607ccb91f840)

- **`rlb-public.nomad`** — internet-facing edge load balancer (`rbxlabs.net`). Only accepts traffic that either comes from RFC1918 address space or carries a trusted origin-auth header issued to the edge proxy (via a Vault-templated secret); the auth header is stripped before the request reaches any backend.
- **`rlb-private.nomad`** — internal-only load balancer (`rbxlabs.dev`), restricted to RFC1918 source addresses only.

Both jobs follow the same pattern:

- **`haproxy-certbot` sidecar** — a `certbot` prestart sidecar that solves an NS1 DNS-01 challenge (credentials pulled from Vault) to obtain/renew a wildcard Let's Encrypt certificate, and concatenates the full chain + private key into a single PEM (`full.pem`) HAProxy can consume directly.
- **`haproxy` task** — runs the `roblox-load-balancer` Docker image with `--configuration-file-path` pointing at a Nomad `template`-rendered RLB config, and:
  - Downloads shared static assets (error pages, a 1×1 tracking gif) from S3 via a Nomad `artifact` block.
  - Renders both the RLB daemon config (`local/haproxy.conf.yml`) and the base HAProxy template (`local/haproxy.cfg.tmpl`) as Nomad `template` stanzas.
  - Registers `reload_on_changes_detected_for_files` against the certbot-managed cert bundle, so certificate renewals trigger an automatic HAProxy reload with zero manual intervention.
  - Exposes `stats` (HAProxy stats page, `:8084`) and `prometheus` (`:9101`) ports, both restricted to private address ranges, plus public `http`/`https` (`:80`/`:443`).
  - Serves `/_/_/health` for load-balancer/orchestrator health checks and `/_/_/1px.gif` as a lightweight beacon endpoint.
  - Normalizes duplicate/trailing slashes in request paths before any ACL matching occurs.
  - Falls back to a `default.http` backend that returns a themed 404 for any host/path that doesn't match a registered service.

These jobs are a good starting point for adapting RLB to your own Nomad + Consul + Vault environment, but the RLB daemon itself has no direct dependency on Nomad or Vault — it only needs a Consul catalog and a template file.

## Repository layout

```txt
src/
├── main.go                 # Entrypoint: parses config, connects to Consul, starts HAProxy + daemon loop
├── flags/                  # CLI flag definitions and glog setup
├── configuration/          # Config file parsing (YAML/JSON/TOML), defaults, and validation
├── consul/                 # Consul API client initialization
├── services/                # Fetches Consul catalog, parses service tags into ServiceConfig, builds backends/rules
│   └── types/               # Service/Node/ServiceConfig/Frontend/Backend types (with hashing for change detection)
├── haproxy/                 # Template rendering, process supervision (start/reload/teardown), config validation
└── daemon/                  # Main refresh loop, SIGUSR1 handling, and filesystem watcher for cert reloads
```

## Notice

## Usage of Roblox, or any of its assets.

# ***This project is not affiliated with Roblox Corporation.***

The usage of the name Roblox and any of its assets is purely for the purpose of providing a clear understanding of the project's purpose and functionality. This project is not endorsed by Roblox Corporation, and is not intended to be used for any commercial purposes.

Any code in this project was soley produced with or without the assistance of error traces and/or behaviour analysis of public facing APIs.