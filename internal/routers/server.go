package routers

import (
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	
	"github.com/aloys.zy/aloys-webhook-example/internal/setting"
	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	ctrl "sigs.k8s.io/controller-runtime"
)

// serve handles the HTTP portion of a request prior to handing to an admit function.
func serve(w http.ResponseWriter, r *http.Request, admit setting.AdmitHandler) {
	setupLog := ctrl.Log.WithName("server")

	// // 记录请求的基本信息
	// lg.Infow(
	// 	"Received incoming request",
	// 	"method", r.Method,
	// 	"url", r.URL.String(),
	// 	"remoteAddr", r.RemoteAddr,
	// 	"userAgent", r.UserAgent(),
	// )

	// 尝试读取HTTP请求体的内容，并将其存储在 body 变量中。如果读取失败，body 将保持为空。
	var body []byte
	if r.Body != nil {
		if data, err := io.ReadAll(r.Body); err == nil {
			body = data
		} else {
			setupLog.Error(err,
				"Failed to read request body", "stacktrace", string(debug.Stack()))
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
	}

	// 检查请求头中的 Content-Type 是否为 application/json。如果不是，记录错误日志并返回，不继续处理。
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		setupLog.V(1).Info(
			"Unsupported Content-Type",
			"contentType", contentType,
			"expectedContentType", "application/json",
			"method", r.Method,
			"url", r.URL.String(),
			"remoteAddr", r.RemoteAddr,
		)
		http.Error(w, "Unsupported Content-Type: application/json required", http.StatusUnsupportedMediaType)
		return
	}

	// 使用 UniversalDeserializer 尝试将请求体解码为Kubernetes对象。gvk 是解码后的对象的 GroupVersionKind。
	deserializer := setting.Codecs.UniversalDeserializer()
	obj, gvk, err := deserializer.Decode(body, nil, nil)
	if err != nil {
		// 如果解码失败，记录错误日志并返回HTTP 400 Bad Request
		setupLog.Error(err,
			"Failed to decode request",
			"body", string(body),
			"method", r.Method,
			"url", r.URL.String(),
			"remoteAddr", r.RemoteAddr,
		)
		http.Error(w, fmt.Sprintf("Request could not be decoded: %v", err), http.StatusBadRequest)
		return
	}

	var responseObj runtime.Object
	// 处理不同API版本的请求
	switch *gvk {
	case v1beta1.SchemeGroupVersion.WithKind("AdmissionReview"):
		// 将解码后的对象转换为 v1beta1.AdmissionReview 类型。
		requestedAdmissionReview, ok := obj.(*v1beta1.AdmissionReview)
		if !ok {
			setupLog.V(1).Info(
				"Expected v1beta1.AdmissionReview but got different type",
				"expectedType", "*admissionv1beta1.AdmissionReview",
				"actualType", fmt.Sprintf("%T", obj),
				"method", r.Method,
				"url", r.URL.String(),
				"remoteAddr", r.RemoteAddr,
			)
			http.Error(w, "Unexpected object type for v1beta1.AdmissionReview", http.StatusBadRequest)
			return
		}

		// 创建一个新的 v1beta1.AdmissionReview 对象作为响应。
		responseAdmissionReview := &v1beta1.AdmissionReview{}
		responseAdmissionReview.SetGroupVersionKind(*gvk)
		responseAdmissionReview.Response = admit.V1beta1(*requestedAdmissionReview)
		responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID
		responseObj = responseAdmissionReview

	case admissionv1.SchemeGroupVersion.WithKind("AdmissionReview"):
		// 将解码后的对象转换为 admissionv1.AdmissionReview 类型。
		requestedAdmissionReview, ok := obj.(*admissionv1.AdmissionReview)
		if !ok {
			setupLog.V(1).Info(
				"Expected admissionv1.AdmissionReview but got different type",
				"expectedType", "*admissionv1.AdmissionReview",
				"actualType", fmt.Sprintf("%T", obj),
				"method", r.Method,
				"url", r.URL.String(),
				"remoteAddr", r.RemoteAddr,
			)
			http.Error(w, "Unexpected object type for admissionv1.AdmissionReview", http.StatusBadRequest)
			return
		}

		// 创建一个新的 admissionv1.AdmissionReview 对象作为响应。
		responseAdmissionReview := &admissionv1.AdmissionReview{}
		responseAdmissionReview.SetGroupVersionKind(*gvk)
		responseAdmissionReview.Response = admit.V1(*requestedAdmissionReview)
		responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID
		responseObj = responseAdmissionReview

	default:
		// 如果请求的 GroupVersionKind 不是 v1beta1 或 v1，则记录错误日志并返回HTTP 400 Bad Request
		setupLog.Info(
			"Unsupported group version kind",
			"groupVersionKind", gvk,
			"method", r.Method,
			"url", r.URL.String(),
			"remoteAddr", r.RemoteAddr,
		)
		http.Error(w, fmt.Sprintf("Unsupported group version kind: %v", gvk), http.StatusBadRequest)
		return
	}

	// 将响应对象序列化为JSON格式
	respBytes, err := json.Marshal(responseObj)
	if err != nil {
		// 如果序列化失败，记录错误日志并返回HTTP 500 Internal Server Error。
		setupLog.Error(err, "Failed to marshal response",
			"responseObject", responseObj,
			"method", r.Method,
			"url", r.URL.String(),
			"remoteAddr", r.RemoteAddr,
		)
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	// 设置响应头为 application/json
	w.Header().Set("Content-Type", "application/json")
	// 将序列化后的JSON数据写入HTTP响应。
	if _, err := w.Write(respBytes); err != nil {
		setupLog.Error(err, "Failed to write response",
			"method", r.Method,
			"url", r.URL.String(),
			"remoteAddr", r.RemoteAddr,
		)
	}
}
