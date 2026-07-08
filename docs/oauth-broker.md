# Clockr OAuth Broker

The OAuth broker is a separate deployable Go binary:

```bash
go run ./cmd/oauth-broker
```

It keeps the Google Web OAuth client secret in the server environment and stores
only short-lived coordination records in the configured datastore. It does not
create durable tables for Google access tokens, Google refresh tokens, or
Calendar event data.

## Environment

Required:

- `CLOCKR_BROKER_PUBLIC_ORIGIN`: public HTTPS origin, for example
  `https://auth.clockr.app`.
- `CLOCKR_BROKER_GOOGLE_CLIENT_ID`: Google Web OAuth client id.
- `CLOCKR_BROKER_GOOGLE_CLIENT_SECRET`: Google Web OAuth client secret.
- `CLOCKR_BROKER_DATASTORE_DSN`: SQLite DSN for the broker datastore.

Optional:

- `CLOCKR_BROKER_LISTEN_ADDR`: listen address, default `:8080`.
- `CLOCKR_BROKER_DESKTOP_HANDOFF_URL`: desktop handoff URL, default
  `clockr://oauth/google/handoff`.
- `CLOCKR_BROKER_STATE_TTL`: OAuth state TTL, default `5m`, maximum `10m`.
- `CLOCKR_BROKER_HANDOFF_TTL`: handoff TTL, default `2m`, maximum `5m`.
- `CLOCKR_BROKER_GOOGLE_SCOPES`: space- or comma-separated Google scopes,
  default `https://www.googleapis.com/auth/calendar.readonly`.

## Deployment Notes

- HTTPS/domain: provision `auth.clockr.app` with HTTPS and HSTS. Configure the
  Google Web OAuth redirect URI as
  `https://auth.clockr.app/v1/google/oauth/callback`.
- Secret management: inject the Google client id and client secret from a
  managed secret store. Restrict runtime access and keep audit logs for reads
  and rotations.
- Datastore: start with a small SQLite database on durable storage for the first
  deployable broker. The schema contains OAuth state and handoff coordination
  records with expiry and one-time-use fields only.
- Logging: keep authorization codes, handoff codes, Google access tokens, and
  Google refresh tokens out of logs, metrics labels, traces, and error messages.
- Metrics: track `/healthz`, `/readyz`, start attempts, datastore failures,
  callback/handoff/refresh/revoke outcomes, expiry counts, and one-time-use
  replay attempts.
- Operational ownership: assign an owner for deploys, uptime, datastore backups,
  secret rotation, Google OAuth consent-screen health, quota alerts, and
  emergency disablement.

## Current Service Surface

- `GET /healthz`: process health.
- `GET /readyz`: validates configuration and checks datastore connectivity.
- `POST /v1/google/oauth/start`: creates a short-lived broker state and returns
  a Google authorization URL.
- `GET /v1/google/oauth/callback`: reserved for DYL-82.
- `POST /v1/google/oauth/handoff`: reserved for DYL-82.
- `POST /v1/google/oauth/refresh`: reserved for broker refresh work.
- `POST /v1/google/oauth/revoke`: reserved for broker disconnect/revoke work.
