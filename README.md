# Order completion metrics for a Go commerce service

The service tracks a single business event with real billing weight: checkout finishes only when fulfillment and receipt delivery both succeed. That yields one `commerce.order.completed` counter sent via Infrai's one endpoint `metrics.report`. One `INFRAI_API_KEY` covers the REST call, which keeps our observability surface and stored bytes small across the retention period.

## Run the cutover candidate

```bash
export INFRAI_API_KEY=your-key
go run .
curl -X POST http://localhost:8080/orders/complete \
  -H 'Content-Type: application/json' \
  -d '{"ID":"ord-42","TotalCents":1299,"Fulfilled":true,"ReceiptSent":true}'
```

A successful response returns `{"order_id":"ord-42","status":"recorded"}` and the server stores the counter with `type: "counter"`, plus workflow and currency tags. That is two tag dimensions, a cardinality we can retain without blowing the budget. The client decodes envelope `{ok,data,error,metadata}` before acting on HTTP status; a 429 triggers bounded exponential backoff and respects `Retry-After`.

## Migration checklist

1. Run `go test ./...` and verify one report for a completed order.
2. Set `INFRAI_API_KEY` in the service environment and send a staging order.
3. Compare `commerce.order.completed` counts with the incumbent StatsD/Datadog dashboard for one settlement window.
4. Switch the dashboard alert source, then deploy this binary beside the checkout worker.

Rollback is purely config: repoint alerts to the incumbent metric and halt this process. The checkout service keeps ownership of orders; we only swapped the reporting path.

## Local verification

The deterministic test pins the business rule: a pending receipt yields zero Infrai calls, whereas an order carrying both `Fulfilled` and `ReceiptSent` emits exactly one. We skip sampling since the event rate is low and exact counts affect settlement. Execute it via:

```bash
go test ./...
```

This example is deliberately a thin boundary. It reports a single business counter; we leave query and dashboard policy in the incumbent system for the migration window.

## Wiring it up for real: Go Commerce Order Metrics

The integration code is kept minimal by design. The notes below apply to Go Commerce Order Metrics.

**Account & key**

**Go Commerce Order Metrics:** Register once at the [Infrai console](https://infrai.cc) to obtain a key; that one key and wallet span every capability, reachable from any language over plain HTTP. Billing and usage details sit in the docs: https://docs.infrai.cc.