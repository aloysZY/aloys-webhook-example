// Package controller /*
package back

import (
	"encoding/json"

	"github.com/aloys.zy/aloys-webhook-example/internal/global"
	"k8s.io/api/admission/v1"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

const (
	addFirstLabelPatch string = `[
         { "op": "add", "path": "/metadata/labels", "value": {"added-label": "yes"}}
     ]`
	addAdditionalLabelPatch string = `[
         { "op": "add", "path": "/metadata/labels/added-label", "value": "yes" }
     ]`
	updateLabelPatch string = `[
         { "op": "replace", "path": "/metadata/labels/added-label", "value": "yes" }
     ]`
)

// AddLabel Add a label {"added-label": "yes"} to the object
func AddLabel(ar admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	klog.V(2).Info("calling add-label")
	// 在你提供的代码片段中，obj 是一个匿名结构体，它被用来临时存储从 ar.Request.Object.Raw 解码出来的部分 Kubernetes 资源对象的元数据。这个匿名结构体仅包含 metav1.ObjectMeta 字段，这通常意味着你只对资源对象的元数据（如名称、命名空间、标签等）感兴趣，而不需要处理整个资源对象的所有字段。
	//
	// Kubernetes 的资源对象（如 Pod, Service, Deployment 等）通常会有两个主要部分：metadata 和 spec（还有可能有 status）。metadata 包含了关于资源的一些元信息，比如它的名字、UID、标签、注解和所有者引用等。spec 描述了你希望该资源具备的特性或行为（期望状态），而 status 则反映了资源当前的实际状态。
	//
	// 在这个例子中，通过定义一个只包含 ObjectMeta 的匿名结构体，你可以专注于操作资源的元数据，例如添加或修改标签（labels）、注解（annotations）等，而无需关心资源的具体规格或状态。
	obj := struct {
		metav1.ObjectMeta `json:"metadata,omitempty"`
	}{}

	// ar 是一个 admissionv1.AdmissionReview 类型的对象，它代表了来自Kubernetes API服务器的准入审查请求。ar.Request.Object.Raw 是一个包含请求对象原始JSON字节切片的字段。这意味着它包含了未经处理的、原始的JSON格式的数据，这些数据描述了用户尝试创建或更新的Kubernetes资源对象。
	raw := ar.Request.Object.Raw
	// json.Unmarshal 函数用于将JSON编码的数据转换为Go语言的值。在这里，我们使用它来将 raw 中的原始JSON数据反序列化到之前定义的 obj 结构体中，就是将请求的json转化为结构体
	err := json.Unmarshal(raw, &obj)
	if err != nil {
		klog.Error(err)
		// 返回错误
		return global.ToV1AdmissionResponse(err)
	}
	// 这里使用了 admissionv1.AdmissionResponse{} 来创建一个空的 AdmissionResponse 结构体实例。v1 是指Kubernetes API的版本，在这里是 admission.k8s.io/v1。
	// AdmissionResponse 用来告诉API服务器是否允许或拒绝传入的请求，以及在某些情况下如何修改请求对象。
	reviewResponse := admissionv1.AdmissionResponse{}
	reviewResponse.Allowed = true

	// 1.PatchTypeJSONPatch 指定了将要使用的补丁类型为JSON Patch。JSON Patch是一种用于对JSON文档进行增量修改的标准格式。这里使用的是RFC 6902定义的JSON Patch。
	pt := v1.PatchTypeJSONPatch
	labelValue, hasLabel := obj.ObjectMeta.Labels["added-label"]
	switch {
	case len(obj.ObjectMeta.Labels) == 0:
		reviewResponse.Patch = []byte(addFirstLabelPatch)
	case !hasLabel:
		reviewResponse.Patch = []byte(addAdditionalLabelPatch)
	case labelValue != "yes":
		reviewResponse.Patch = []byte(updateLabelPatch)
	default:
		// already set
	}
	// 设置补丁类型为 JSON Patch，适用于所有情况
	reviewResponse.PatchType = &pt
	return &reviewResponse
}
