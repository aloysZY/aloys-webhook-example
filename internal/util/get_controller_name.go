package util

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// GetControllerName 根据 Pod 的 OwnerReferences 查找并记录控制器的相关信息。
// reason 和 message 作为参数传递，用于记录事件。
// 只返回一个错误。
func GetControllerName(pod *corev1.Pod, reason, message string) error {
	setupLog := ctrl.Log.WithName("GetControllerName")

	// 获取 Pod 名称，优先使用 pod.Name，如果为空则使用 GenerateName
	podName := pod.Name
	if podName == "" {
		podName = pod.GenerateName
	}
	if podName == "" {
		return fmt.Errorf("pod name is empty")
	}

	// 遍历 OwnerReferences，查找控制器
	for _, owner := range pod.GetOwnerReferences() {
		switch owner.Kind {
		case "ReplicaSet":
			return handleReplicaSet(pod, owner, reason, message)

		case "Job":
			return handleJob(pod, owner, reason, message)

		case "StatefulSet":
			return handleStatefulSet(pod, owner, reason, message)

		default:
			setupLog.V(1).Info("Logged event for unknown controller", "kind", owner.Kind, "controller", owner.Name)
			eventRecorder.Eventf(
				pod,
				corev1.EventTypeNormal,
				reason,
				message,
			)
			return nil
		}
	}

	return fmt.Errorf("no suitable controller found in OwnerReferences")
}

// 辅助函数：处理 ReplicaSet
func handleReplicaSet(pod *corev1.Pod, owner metav1.OwnerReference, reason, message string) error {
	rs, err := clientSet.AppsV1().ReplicaSets(pod.Namespace).Get(context.Background(), owner.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get ReplicaSet: %w", err)
	}

	// 查找 ReplicaSet 的 OwnerReference，确认是否为 Deployment
	for _, rsOwner := range rs.OwnerReferences {
		if rsOwner.Kind == "Deployment" {
			return handleDeployment(pod, rsOwner, reason, message)
		}
	}

	// 如果没有找到 Deployment，记录 ReplicaSet 事件
	eventRecorder.Eventf(
		rs,
		corev1.EventTypeNormal,
		reason,
		message,
	)
	return nil
}

// 辅助函数：处理 Deployment
func handleDeployment(pod *corev1.Pod, owner metav1.OwnerReference, reason, message string) error {
	parts := strings.Split(owner.Name, "-")
	if len(parts) == 0 {
		return fmt.Errorf("failed to split Deployment name: %s", owner.Name)
	}

	deploymentName := parts[0]

	dep, err := clientSet.AppsV1().Deployments(pod.Namespace).Get(context.Background(), deploymentName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get Deployment: %w", err)
	}

	// 记录事件
	eventRecorder.Eventf(
		dep,
		corev1.EventTypeNormal,
		reason,
		message,
	)
	return nil
}

// 辅助函数：处理 Job
func handleJob(pod *corev1.Pod, owner metav1.OwnerReference, reason, message string) error {
	job, err := clientSet.BatchV1().Jobs(pod.Namespace).Get(context.Background(), owner.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get Job: %w", err)
	}
	// 记录事件
	eventRecorder.Eventf(
		job,
		corev1.EventTypeNormal,
		reason,
		message,
	)
	return nil
}

// 辅助函数：处理 StatefulSet
func handleStatefulSet(pod *corev1.Pod, owner metav1.OwnerReference, reason, message string) error {
	parts := strings.Split(owner.Name, "-")
	if len(parts) < 2 {
		return fmt.Errorf("failed to split StatefulSet name: %s", owner.Name)
	}

	statefulSetName := strings.Join(parts[:2], "-")

	sts, err := clientSet.AppsV1().StatefulSets(pod.Namespace).Get(context.Background(), statefulSetName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get StatefulSet: %w", err)
	}

	// 记录事件
	eventRecorder.Eventf(
		sts,
		corev1.EventTypeNormal,
		reason,
		message,
	)
	return nil
}
