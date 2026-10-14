package main

import "time"

type Order struct {
	ID          string
	TotalCents  int
	Fulfilled   bool
	ReceiptSent bool
}

// ReportOrderMetrics emits counters only after the business state is complete.
func ReportOrderMetrics(client *MetricsClient, order Order, now time.Time) error {
	if !order.Fulfilled || !order.ReceiptSent {
		return nil
	}
	tags := map[string]string{"workflow": "checkout_fulfillment", "currency": "USD"}
	return client.Report(map[string]any{
		"name":      "commerce.order.completed",
		"value":     1,
		"type":      "counter",
		"tags":      tags,
		"timestamp": now.UTC().Format(time.RFC3339),
	})
}
