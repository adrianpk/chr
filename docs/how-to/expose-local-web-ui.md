# Expose the Local Web UI Deliberately

The Web UI is local by default. Running `chr web serve` binds to
`127.0.0.1:7331` unless the project config or command flags say otherwise.

Use a non-loopback bind only when another system is responsible for network
reachability, firewalling, and TLS termination.

## Local Only

```sh
chr web serve
```

Expected boundary:

- bind: `127.0.0.1`
- token requirement: disabled in `auto` mode
- origins: same-origin only
- exposure: local browser on the same machine

## Trusted Internal Network

For password-protected navigation, configure
[single-user login](configure-web-login.md), then use
`chr web serve --allow-lan --require-login`. The trusted-LAN mode below describes
access without a configured login.

For direct browser access from another computer on a trusted internal network:

```sh
chr web serve --allow-lan
```

Open the development server's reachable hostname or address from the browser
computer. For example, if the server is named `devbox`:

```text
http://devbox:7331/
```

Without a configured login, this mode requires no token, token environment variable, authentication cookie
or custom browser header. Pages, links, refreshes, HTMX updates and forms use
the same resolved policy. With default settings it binds to `0.0.0.0:7331`.
An existing configured network interface is retained; explicit `--host` and
`--port` override the bind as usual.

The product default is `false`. Retain this choice for a project in its
`.chr/default/config.yml`:

```yml
web:
  allow_lan: true
```

Then `chr web serve` uses LAN mode without repeating the flag. An explicit
`--allow-lan=false` overrides this setting and preserves normal host/token
configuration resolution; with default host settings it restores the loopback
bind. A configured or explicit `true`/`always` token requirement
conflicts with LAN mode. Select `--require-token=false` explicitly to override
that configured requirement, or keep the token-protected mode.

Startup and `--dry-run` output show `allow_lan`, the bind and `token_required`.
LAN mode retains origin checks, write confirmations and revision checks. It
does not configure a firewall or filter incoming connections by private CIDR.

## Token-Protected Bind

A configured single-user login takes precedence over this legacy token mode.
Tokens do not bypass login.

Set a token before starting the server:

```sh
export CHR_WEB_TOKEN='replace-with-a-long-random-value'
chr web serve --host 0.0.0.0 --port 7331
```

Preview the resolved exposure boundary first:

```sh
CHR_WEB_TOKEN='replace-with-a-long-random-value' \
  chr web serve --host 0.0.0.0 --port 7331 --dry-run
```

In `auto` mode, non-loopback binds require the token. The token can be supplied
with one of these request forms:

- `X-CHR-Web-Token: <token>`
- `Authorization: Bearer <token>`
- `?token=<token>`

The query form exists for simple browser access through private links. Avoid
sharing URLs that contain the token. A query token authorizes that request;
browser links do not inherit it. Use `--allow-lan` for token-free navigation on
a trusted internal network.

## Allowed Origins

Cross-origin browser requests are denied by default. Configure exact origins
when a private proxy or tunnel serves the Web UI from another origin:

```yml
web:
  host: 0.0.0.0
  port: 7331
  require_token: auto
  token_env: CHR_WEB_TOKEN
  allowed_origins: "https://chr.example.internal"
```

Allowed origins must be exact origins. Wildcards are not supported.

## Operator-Owned Boundary

Chryso does not configure or manage Tailscale, Cloudflare Tunnel, reverse
proxies, firewall rules, DNS, certificates, or public hosting. Keep those
systems outside `chr` and expose the Web UI only through infrastructure you
control.

Startup and request logs do not print token values or query strings.

See [Run the local Web UI](run-local-web-ui.md) for localhost-only operation.
