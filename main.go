package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func main() {
	client, err := NewMetricsClient(nil)
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/orders/complete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var order Order
		if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
			http.Error(w, "invalid order", http.StatusBadRequest)
			return
		}
		if err := ReportOrderMetrics(client, order, time.Now()); err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "recorded", "order_id": order.ID})
	})
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
