package util

import (
	"github.com/aloys.zy/aloys-webhook-example/internal/logger"
	corev1 "k8s.io/api/core/v1"
)

// UpdateAnnotationForInvalidLabel 更新节点的注解为指定的值，如果注解不存在或不是指定的值则修改
//
// 参数:
// - node: 需要更新注解的节点对象
// - key: 注解的键
// - value: 注解的值
func UpdateAnnotationForInvalidLabel(node *corev1.Node, key, value string) {
	sugaredLogger := logger.WithName("util.UpdateAnnotationForInvalidLabel")

	// 确保注解存在
	if node.Annotations == nil {
		node.Annotations = make(map[string]string)
	}

	// 获取现有的注解值
	currentAnnotation, exists := node.Annotations[key]

	// 如果注解不存在或者不是指定的值，则更新为指定的值
	if !exists || currentAnnotation != value {
		node.Annotations[key] = value
		sugaredLogger.Infof("Updated or added annotation %s to %s for node %s", key, value, node.Name)
	} else {
		sugaredLogger.Debugf("Annotation %s already set to %s for node %s, no update needed", key, value, node.Name)
	}
}
