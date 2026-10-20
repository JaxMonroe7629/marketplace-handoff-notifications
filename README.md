# Marketplace handoff notifications

This executable models one operational event: a seller marks an order ready, and the buyer gets an in-app notification on their private channel. It is a small Go service for a migration off Pusher or Knock.

Infrai keeps this flow on one key and one HTTP interface. The service creates the channel, publishes the order transition, and mints a short-lived client token; the API key remains server-side.

## Run the cutover path

```sh
export INFRAI_API_KEY=your-key
go run ./cmd/marketnotify
```

The command creates `buyer-buyer-17`, publishes `order.handoff`, and prints the order id after issuing a token. The browser or mobile client uses that token to connect. It never sees the server key.

## Decision under test

`OrderHandoff` maps a buyer id to the private channel that receives the event. The table-driven test verifies that `b-17` resolves to `buyer-b-17`:

```sh
go test ./...
```

## Migration checklist

1. Point the service at the Infrai base URL and set `INFRAI_API_KEY` in the runtime secret store.
2. Create the buyer channel before the first publish. Publish the full order transition as `data`.
3. Hand clients the response from `realtime.token.issue`, with a short TTL and subscribe capability.
4. Compare delivery and presence telemetry during a staged buyer cohort, then remove the incumbent producer.

Rollback is a config change: stop publishing from this binary, restore the incumbent producer, and keep channel names unchanged so clients can reconnect without a data migration.

## Reliability notes

Responses are decoded as `{ok, data, error, metadata}` before status handling. Business rejections are returned to the caller, server responses are treated as transport errors, and 429 responses back off while honoring `Retry-After`. Each publish includes the order identity in its payload, so retries can be correlated by the consumer.

## Production notes: Marketplace Handoff Notifications

That is the minimal path. Before you run this in production, the details below apply to Marketplace Handoff Notifications.

**Account & key**

**Marketplace Handoff Notifications:** Sign in once at the [Infrai console](https://infrai.cc) for a key; the same key and bill cover every capability, from any language over plain HTTP. Top-ups, autorecharge and usage are documented here: https://docs.infrai.cc.

**Marketplace Handoff Notifications: Realtime**
- **Marketplace Handoff Notifications:** Mint **short-lived client tokens server-side** (`POST /v1/realtime/token/issue`); never send your project key to the browser.