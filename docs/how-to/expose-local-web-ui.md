# Expose the Local Web UI Deliberately

The Web UI is local by default. Running `chc web serve` binds to
`127.0.0.1:7331` unless the project config or command flags say otherwise.

Use a non-loopback bind only when another system is responsible for network
reachability, firewalling, and TLS termination.

## Local Only

```sh
chc web serve
```

Expected boundary:

- bind: `127.0.0.1`
- token requirement: disabled in `auto` mode
- origins: same-origin only
- exposure: local browser on the same machine

## Private Network Or Tunnel

Set a token before starting the server:

```sh
export CHC_WEB_TOKEN='replace-with-a-long-random-value'
chc web serve --host 0.0.0.0 --port 7331
```

In `auto` mode, non-loopback binds require the token. The token can be supplied
with one of these request forms:

- `X-CHC-Web-Token: <token>`
- `Authorization: Bearer <token>`
- `?token=<token>`

The query form exists for simple browser access through private links. Avoid
sharing URLs that contain the token.

## Allowed Origins

Cross-origin browser requests are denied by default. Configure exact origins
when a private proxy or tunnel serves the Web UI from another origin:

```yml
web:
  host: 0.0.0.0
  port: 7331
  require_token: auto
  token_env: CHC_WEB_TOKEN
  allowed_origins: "https://chc.example.internal"
```

Allowed origins must be exact origins. Wildcards are not supported.

## Operator-Owned Boundary

ChrysoCoord does not configure or manage Tailscale, Cloudflare Tunnel, reverse
proxies, firewall rules, DNS, certificates, or public hosting. Keep those
systems outside `chc` and expose the Web UI only through infrastructure you
control.

Startup and request logs do not print token values or query strings.
