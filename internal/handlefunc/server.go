package handlefunc

import (
	"fmt"
	"io"
	"net/http"

	"github.com/aloys.zy/aloys-webhook-example/api"
	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/klog/v2"
)

// serve handles the http portion of a request prior to handing to an admit
// function
// w http.ResponseWriter：HTTP响应写入器，用于发送响应给客户端。
// r *http.Request：HTTP请求对象，包含了请求的所有信息。
// admit admitHandler：一个实现了不同API版本处理逻辑的处理器，通过 newDelegateToV1AdmitHandler 创建
func serve(w http.ResponseWriter, r *http.Request, admit admitHandler) {
	// 这里尝试读取HTTP请求体的内容，并将其存储在 body 变量中。如果读取失败，body 将保持为空。
	var body []byte
	if r.Body != nil {
		if data, err := io.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	// verify the content type is accurate
	// 检查请求头中的 Content-Type 是否为 application/json。如果不是，记录错误日志并返回，不继续处理。
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		klog.Errorf("contentType=%s, expect application/json", contentType)
		return
	}

	klog.V(2).Info(fmt.Sprintf("handling request: %s", body))

	// 使用 UniversalDeserializer 尝试将请求体解码为Kubernetes对象。gvk 是解码后的对象的GroupVersionKind。
	deserializer := api.Codecs.UniversalDeserializer()
	obj, gvk, err := deserializer.Decode(body, nil, nil)
	if err != nil {
		// 如果解码失败，记录错误日志并返回HTTP 400 Bad Request
		msg := fmt.Sprintf("Request could not be decoded: %v", err)
		klog.Error(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	var responseObj runtime.Object
	// 处理不同API版本的请求
	switch *gvk {
	case v1beta1.SchemeGroupVersion.WithKind("AdmissionReview"):
		// 将解码后的对象转换为 v1beta1.AdmissionReview 类型。
		requestedAdmissionReview, ok := obj.(*v1beta1.AdmissionReview)
		if !ok {
			klog.Errorf("Expected v1beta1.AdmissionReview but got: %T", obj)
			return
		}
		// 创建一个新的 v1beta1.AdmissionReview 对象作为响应。
		responseAdmissionReview := &v1beta1.AdmissionReview{}
		responseAdmissionReview.SetGroupVersionKind(*gvk)
		responseAdmissionReview.Response = admit.v1beta1(*requestedAdmissionReview)
		// 调用 admit.v1beta1 处理请求，并设置响应的 UID。
		responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID
		// 将响应对象赋值给 responseObj。
		responseObj = responseAdmissionReview
	case admissionv1.SchemeGroupVersion.WithKind("AdmissionReview"):
		requestedAdmissionReview, ok := obj.(*admissionv1.AdmissionReview)
		if !ok {
			klog.Errorf("Expected admissionv1.AdmissionReview but got: %T", obj)
			return
		}
		responseAdmissionReview := &admissionv1.AdmissionReview{}
		responseAdmissionReview.SetGroupVersionKind(*gvk)
		responseAdmissionReview.Response = admit.v1(*requestedAdmissionReview)
		responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID
		responseObj = responseAdmissionReview
	default:
		// 如果请求的 GroupVersionKind 不是 v1beta1 或 v1，则记录错误日志并返回HTTP 400 Bad Request
		msg := fmt.Sprintf("Unsupported group version kind: %v", gvk)
		klog.Error(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	klog.V(2).Info(fmt.Sprintf("sending response: %v", responseObj))
	// 将响应对象序列化为JSON格式
	respBytes, err := json.Marshal(responseObj)
	if err != nil {
		// 如果序列化失败，记录错误日志并返回HTTP 500 Internal Server Error。
		klog.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// 设置响应头为 application/json
	w.Header().Set("Content-Type", "application/json")
	// 将序列化后的JSON数据写入HTTP响应。
	if _, err := w.Write(respBytes); err != nil {
		klog.Error(err)
	}
}
