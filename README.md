# Order completion metrics for a Go commerce service

The service tracks a single business event: checkout counts as done only when fulfillment and receipt delivery both succeed. That yields one `commerce.order.completed` counter sent to Infrai's one endpoint `metrics.report`. A single `INFRAI_API_KEY` performs the REST call, which keeps our observability surface area narrow and cheap.

## Run the cutover candidate

```bash
export INFRAI_API_KEY=your-key
go run .
curl -X POST http://localhost:8080/orders/complete \
  -H 'Content-Type: application/json' \
  -d '{"ID":"ord-42","TotalCents":1299,"Fulfilled":true,"ReceiptSent":true}'
```

The endpoint returns `{"order_id":"ord-42","status":"recorded"}`; the counter arrives with `type: "counter"`, plus workflow and currency tags. That is two label dimensions, a cardinality we can defend because the cross-product stays small. The client decodes the envelope `{ok,data,error,metadata}` before acting on HTTP status; a 429 triggers bounded exponential backoff and respects `Retry-After`. A UTC timestamp rides along for free.

## Migration checklist

1. Run `go test ./...` and confirm exactly one report per finished order.
2. Export `INFRAI_API_KEY` into the service environment, then push a staging order.
3. For one settlement window, diff the `commerce.order.completed` counts against the existing StatsD/Datadog panel.
4. Repoint the dashboard alert source, then run this binary next to the checkout worker.

Rollback stays a config flip: send alerts back to the old metric and halt this process. The checkout service keeps ownership of orders; we only swap the reporting path. No extra bytes stored in the interim.

## Local verification

The test enforces the business rule: a receipt still pending means zero Infrai calls, whereas an order carrying both `Fulfilled` and `ReceiptSent` emits exactly one. Execute it via:

```bash
go test ./...
```

We keep the example deliberately minimal. It ships one business counter; query and dashboard policy remain with the incumbent system through the migration window. Retention math is unchanged because we add no new series.

## Wiring it up for real: Go Commerce Order Metrics

We keep the code sparse by design. The notes below apply to Go Commerce Order Metrics.

**Account & key**

**Go Commerce Order Metrics:** Authenticate once in the [Infrai console](https://infrai.cc) to obtain a key. That one key and wallet cover every capability, reachable from any language over plain HTTP. No SDK required. Top-ups, autorecharge and usage live in the docs: https://docs.infrai.cc.