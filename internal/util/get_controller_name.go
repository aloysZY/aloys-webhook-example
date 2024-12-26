package util

import (
	"fmt"
	"strings"

	"github.com/aloys.zy/aloys-webhook-example/internal/logger"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetControllerName 根据 Pod 的 OwnerReferences 查找并记录控制器的相关信息。
// reason 和 message 作为参数传递，用于记录事件。
// 只返回一个错误。
func GetControllerName(pod *corev1.Pod, reason, message string) error {
	lg := logger.WithName("GetControllerName")
	clientSet := GetClientSet()
	// 获取 Pod 名称，优先使用 pod.Name，如果为空则使用 GenerateName
	podName := pod.Name
	if podName == "" {
		podName = pod.GenerateName
	}

	// 遍历 OwnerReferences，查找控制器
	for _, owner := range pod.GetOwnerReferences() {
		switch owner.Kind {
		case "ReplicaSet":
			// ReplicaSet 通常是 Deployment 创建的，需要进一步查找 Deployment
			rs, err := clientSet.AppsV1().ReplicaSets(pod.Namespace).Get(context.Background(), owner.Name, metav1.GetOptions{})
			if err != nil {
				lg.Error("Failed to get ReplicaSet", zap.Error(err), zap.String("replicaSet", owner.Name))
				return fmt.Errorf("failed to get ReplicaSet: %v", err)
			}

			// 查找 ReplicaSet 的 OwnerReference，确认是否为 Deployment
			for _, rsOwner := range rs.OwnerReferences {
				if rsOwner.Kind == "Deployment" {
					parts := strings.Split(rsOwner.Name, "-")
					if len(parts) > 0 {
						deploymentName := parts[0]

						// 获取 Deployment
						dep, err := clientSet.AppsV1().Deployments(pod.Namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
						if err != nil {
							lg.Error("Failed to get Deployment", zap.Error(err), zap.String("deployment", deploymentName))
							return fmt.Errorf("failed to get Deployment: %v", err)
						}

						// 记录事件
						EventRecorder().Eventf(
							dep,
							corev1.EventTypeNormal,
							reason,
							message,
						)

						lg.Info("Logged event for Deployment", zap.String("deployment", dep.Name))
						return nil
					} else {
						lg.Warn("Failed to split Deployment name", zap.String("name", rsOwner.Name))
						return fmt.Errorf("failed to split Deployment name: %s", rsOwner.Name)
					}
				}
			}

			// 如果没有找到 Deployment，记录 ReplicaSet 事件
			parts := strings.Split(owner.Name, "-")
			if len(parts) > 0 {
				// replicaSetName := parts[0]

				// 记录事件
				EventRecorder().Eventf(
					rs,
					corev1.EventTypeNormal,
					reason,
					message,
				)

				lg.Info("Logged event for ReplicaSet", zap.String("replicaSet", rs.Name))
				return nil
			} else {
				lg.Warn("Failed to split ReplicaSet name", zap.String("name", owner.Name))
				return fmt.Errorf("failed to split ReplicaSet name: %s", owner.Name)
			}

		case "Job":
			jobName := owner.Name
			// 获取 Job
			job, err := clientSet.BatchV1().Jobs(pod.Namespace).Get(context.TODO(), jobName, metav1.GetOptions{})
			if err != nil {
				lg.Error("Failed to get Job", zap.Error(err), zap.String("job", jobName))
				return fmt.Errorf("failed to get Job: %v", err)
			}

			// 记录事件
			EventRecorder().Eventf(
				job,
				corev1.EventTypeNormal,
				reason,
				message,
			)

			lg.Info("Logged event for Job", zap.String("job", job.Name))
			return nil

		case "StatefulSet":
			parts := strings.Split(owner.Name, "-")
			if len(parts) > 0 {
				statefulSetName := parts[0] + "-" + parts[1]

				// 获取 StatefulSet
				sts, err := clientSet.AppsV1().StatefulSets(pod.Namespace).Get(context.TODO(), statefulSetName, metav1.GetOptions{})
				if err != nil {
					lg.Error("Failed to get StatefulSet", zap.Error(err), zap.String("statefulSet", statefulSetName))
					return fmt.Errorf("failed to get StatefulSet: %v", err)
				}

				// 记录事件
				EventRecorder().Eventf(
					sts,
					corev1.EventTypeNormal,
					reason,
					message,
				)

				lg.Info("Logged event for StatefulSet", zap.String("statefulSet", sts.Name))
				return nil
			} else {
				lg.Warn("Failed to split StatefulSet name", zap.String("name", owner.Name))
				return fmt.Errorf("failed to split StatefulSet name: %s", owner.Name)
			}

		default:
			// 记录事件
			EventRecorder().Eventf(
				pod,
				corev1.EventTypeNormal,
				reason,
				message,
			)

			lg.Warn("Logged event for unknown controller", zap.String("kind", owner.Kind), zap.String("controller", owner.Name))
			return nil

		}
	}
	lg.Info("No suitable controller found in OwnerReferences")
	return fmt.Errorf("no suitable controller found in OwnerReferences")
}
