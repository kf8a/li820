package li820

import (
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
)

func main() {
	go readLicor()

	http.Handle("/metrics", prometheus.Handler())
	http.ListenAndServe(":9092", nil)
}
