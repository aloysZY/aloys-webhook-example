package back

//
// import (
// 	"encoding/json"
// 	"strconv"
//
// 	"github.com/aloys.zy/aloys-webhook-example/api"
// 	"github.com/aloys.zy/aloys-webhook-example/internal/logger"
// 	"google.golang.org/appengine/log"
// 	admissionv1 "k8s.io/api/admission/v1"
// 	corev1 "k8s.io/api/core/v1"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// )
//
// // MutateNodes 处理节点的 AdmissionReview 请求
// func MutateNodes(ar admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
//
// 	sugaredLogger := logger.WithName("mutating nodes.")
//
// 	// 定义节点资源
// 	nodeResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "nodes"}
//
// 	if ar.Request.Resource != nodeResource {
// 		sugaredLogger.Errorf("expect resource to be %s", nodeResource)
// 		return nil
// 	}
// 	raw := ar.Request.Object.Raw
// 	node := corev1.Node{}
// 	deserializer := api.Codecs.UniversalDeserializer()
// 	if _, _, err := deserializer.Decode(raw, nil, &node); err != nil {
// 		sugaredLogger.Error(err)
// 		return api.ToV1AdmissionResponse(err)
// 	}
//
// 	reviewResponse := admissionv1.AdmissionResponse{}
// 	reviewResponse.Allowed = true
// 	// 检查节点是否具有特定的标签
// 	// 构建 JSON Patch
// 	var patches []api.PatchOperation
// 	if newAllocatableCPU, ok := calculateNewAllocatableCPU(&node); ok {
// 		patchOps := append(patches,
// 			api.GetPatchItem("replace", "/status/allocatable/cpu", newAllocatableCPU),
// 			api.GetPatchItem("replace", "/metadata/annotations/node-oversold-cpu", "true"))
// 		patchBytes, _ := json.Marshal(patchOps) // json序列号
// 		reviewResponse.Patch = patchBytes       // 修改的内容添加到path,要json序列号后
// 		sugaredLogger.Info("replace allocatable cpu and annotation.")
// 	} else {
// 		patchOps := append(patches, api.GetPatchItem("add", "/metadata/annotations/node-oversold-cpu", "false"))
// 		patchBytes, _ := json.Marshal(patchOps)
// 		reviewResponse.Patch = patchBytes
// 		sugaredLogger.Info("replace annotation.")
// 	}
// 	pt := admissionv1.PatchTypeJSONPatch
// 	reviewResponse.PatchType = &pt
// 	return &reviewResponse
// }
//
// // calculateNewAllocatableCPU 检查节点是否有特定的标签，并计算新的 allocatable.cpu 值
// func calculateNewAllocatableCPU(node *corev1.Node) (string, bool) {
// 	labels := node.GetLabels()
// 	if labels == nil {
// 		return "", false
// 	}
//
// 	oversoldCPU, ok := labels["node-oversold-cpu"]
// 	if !ok {
// 		return "", false
// 	}
//
// 	// 解析倍数
// 	multiplier, err := strconv.Atoi(oversoldCPU)
// 	if err != nil {
// 		log.Errorf("Error parsing multiplier from label: %v", err)
// 		return "", false
// 	}
//
// 	// 获取当前的 allocatable.cpu 值
// 	currentCPU, err := strconv.ParseFloat(node.Status.Allocatable.Cpu().String(), 64)
// 	if err != nil {
// 		log.Errorf("Error parsing current CPU value: %v", err)
// 		return "", false
// 	}
//
// 	// 计算新的 allocatable.cpu 值  * 1000将单位m转化为cpu
// 	newCPU := currentCPU * float64(multiplier) * 1000
//
// 	// 格式化为字符串
// 	newCPUFormatted := strconv.FormatFloat(newCPU, 'f', -1, 64) + "m"
//
// 	return newCPUFormatted, true
// }
