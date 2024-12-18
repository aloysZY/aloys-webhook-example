package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// 定义并注册自定义指标
var (
	requestCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "webhook_requests_total",
			Help: "Total number of webhook requests.",
		},
		[]string{"path", "status"},
	)
	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "webhook_request_duration_seconds",
			Help:    "Duration of webhook requests in seconds.",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 10), // 从1ms开始，以2为基数，共10个桶
		},
		[]string{"path"},
	)
)

// 自定义ResponseWriter以捕获状态码
type responseCaptureWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseCaptureWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// WithMetrics 包装函数，用于更新自定义指标
func WithMetrics(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		path := req.URL.Path
		start := time.Now()
		rcw := &responseCaptureWriter{ResponseWriter: w, statusCode: http.StatusOK}
		defer func() {
			duration := time.Since(start).Seconds()
			requestDuration.WithLabelValues(path).Observe(duration)
			requestCounter.WithLabelValues(path, strconv.Itoa(rcw.statusCode)).Inc()
		}()

		// 调用原始处理函数
		next.ServeHTTP(rcw, req)
	}
}
