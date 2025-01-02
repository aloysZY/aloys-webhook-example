package setting

import (
	"github.com/aloys.zy/aloys-webhook-example/logger"
	"go.uber.org/zap"
	"gomodules.xyz/jsonpatch/v2"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
)

// GeneratePatchAndResponse 生成 JSON Patch 并返回 AdmissionResponse
func GeneratePatchAndResponse(originalObj, modifiedObj runtime.Object, allowed bool, warning, message string) *admissionv1.AdmissionResponse {
	lg := logger.WithName("util.GeneratePatchAndResponse")

	if originalObj == nil || modifiedObj == nil {
		return constructAdmissionResponse(allowed, nil, warning, message)
	}

	// 序列化原始节点对象
	originalNodeBytes, err := marshalObject(originalObj)
	if err != nil {
		lg.Error("failed to marshal originalNode to JSON: %v", zap.Error(err))
		return ToV1AdmissionResponse(err)
	}

	// 序列化修改后的节点对象
	modifiedNodeBytes, err := marshalObject(modifiedObj)
	if err != nil {
		lg.Error("failed to marshal modified node to JSON: %v", zap.Error(err))
		return ToV1AdmissionResponse(err)
	}

	// 生成 JSON Patch
	patch, err := createJSONPatch(originalNodeBytes, modifiedNodeBytes)
	if err != nil {
		lg.Error("failed to create JSON patch: %v", zap.Error(err))
		return ToV1AdmissionResponse(err)
	}

	// 序列化 JSON Patch
	patchBytes, err := marshalPatch(patch)
	if err != nil {
		lg.Error("failed to marshal JSON patch: %v", zap.Error(err))
		return ToV1AdmissionResponse(err)
	}

	// 构造 AdmissionResponse
	return constructAdmissionResponse(allowed, patchBytes, warning, message)

}

// marshalObject marshals a Kubernetes resource object to JSON bytes.
func marshalObject(obj runtime.Object) ([]byte, error) {
	return json.Marshal(obj)
}

// // marshalNode 将 Node 对象序列化为 JSON 字节数组
// func marshalNode(node *corev1.Node) ([]byte, error) {
// 	return json.Marshal(node)
// }

// createJSONPatch 生成从原始节点到修改后节点的 JSON Patch
func createJSONPatch(original, modified []byte) ([]jsonpatch.JsonPatchOperation, error) {
	return jsonpatch.CreatePatch(original, modified)
}

// marshalPatch 将 JSON Patch 序列化为字节数组
func marshalPatch(patch []jsonpatch.JsonPatchOperation) ([]byte, error) {
	return json.Marshal(patch)
}

// constructAdmissionResponse 构造并返回 AdmissionResponse
func constructAdmissionResponse(allowed bool, patch []byte, warning, message string) *admissionv1.AdmissionResponse {

	response := admissionv1.AdmissionResponse{}
	response.Allowed = allowed
	if patch != nil {
		response.Patch = patch
		response.PatchType = func() *admissionv1.PatchType {
			pt := admissionv1.PatchTypeJSONPatch
			return &pt
		}()
	}

	// 如果有警告信息，添加到 Warnings 中
	if warning != "" {
		response.Warnings = append(response.Warnings, warning)
	}

	// 如果有错误信息，添加到 Result.Message 中
	if message != "" {
		response.Result = &metav1.Status{
			Message: message,
		}
	}

	return &response
}
