package util

import (
	"github.com/aloys.zy/aloys-webhook-example/internal/logger"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
)

// UpdateAnnotationForInvalidLabel 更新节点的注解为指定的值，如果注解不存在或不是指定的值则修改
//
// 参数:
// - node: 需要更新注解的节点对象
// - key: 注解的键
// - value: 注解的值
func UpdateAnnotationForInvalidLabel(node *corev1.Node, key, value string) {
	lg := logger.WithName("util.UpdateAnnotationForInvalidLabel")

	// 确保注解存在
	if node.Annotations == nil {
		node.Annotations = make(map[string]string)
	}

	// 获取现有的注解值
	currentAnnotation, exists := node.Annotations[key]

	// 如果注解不存在或者不是指定的值，则更新为指定的值
	if !exists || currentAnnotation != value {
		// 更新或添加注解
		node.Annotations[key] = value

		// 记录信息级别的日志
		lg.Info(
			"Updated or added annotation", // 日志消息
			zap.String("key", key),        // 键值对：注解的键
			zap.String("value", value),    // 键值对：注解的值
			zap.String("node", node.Name), // 键值对：节点名称
		)
	} else {
		// 记录调试级别的日志
		lg.Debug(
			"Annotation already set",      // 日志消息
			zap.String("key", key),        // 键值对：注解的键
			zap.String("value", value),    // 键值对：注解的值
			zap.String("node", node.Name), // 键值对：节点名称
		)
	}
}
