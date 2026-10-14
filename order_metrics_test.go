package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestReportOrderMetricsRequiresFulfillmentAndReceipt(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Write([]byte(`{"ok":true,"data":{},"error":null,"metadata":{}}`))
	}))
	defer server.Close()
	client := &MetricsClient{BaseURL: server.URL, APIKey: "test", HTTP: server.Client()}
	for _, order := range []Order{{ID: "pending", Fulfilled: false, ReceiptSent: true}, {ID: "ready", Fulfilled: true, ReceiptSent: true}} {
		if err := ReportOrderMetrics(client, order, time.Unix(0, 0)); err != nil {
			t.Fatal(err)
		}
	}
	if calls != 1 {
		t.Fatalf("expected one report for completed order, got %d", calls)
	}
}
