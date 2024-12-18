package metrics

import (
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/aloys.zy/aloys-webhook-example/internal/logger"
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
	sugaredLogger := logger.WithName("metrics") // 假设全局日志记录器已经初始化

	return func(w http.ResponseWriter, req *http.Request) {
		path := req.URL.Path
		start := time.Now()
		rcw := &responseCaptureWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// // 记录请求的基本信息
		// sugaredLogger.Infow(
		// 	"Received incoming request",
		// 	"method", req.Method,
		// 	"url", req.URL.String(),
		// 	"remoteAddr", req.RemoteAddr,
		// 	"userAgent", req.UserAgent(),
		// 	"path", path,
		// )

		// 捕获并恢复任何 panic，以防止未处理的 panic 导致整个服务崩溃
		defer func() {
			if r := recover(); r != nil {
				sugaredLogger.Errorw(
					"Recovered from panic",
					"error", r,
					"stacktrace", string(debug.Stack()), // 记录堆栈跟踪
					"method", req.Method,
					"url", req.URL.String(),
					"remoteAddr", req.RemoteAddr,
					"userAgent", req.UserAgent(),
					"path", path,
				)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			duration := time.Since(start).Seconds()
			requestDuration.WithLabelValues(path).Observe(duration)
			requestCounter.WithLabelValues(path, strconv.Itoa(rcw.statusCode)).Inc()

			// 记录请求完成的日志
			sugaredLogger.Infow(
				"Request completed",
				"method", req.Method,
				"url", req.URL.String(),
				"remoteAddr", req.RemoteAddr,
				"userAgent", req.UserAgent(),
				"path", path,
				"status", rcw.statusCode,
				"duration", duration,
			)
		}()

		// 调用原始处理函数
		next.ServeHTTP(rcw, req)
	}
}
