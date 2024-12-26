package util

import (
	"os"
	"path/filepath"

	"github.com/aloys.zy/aloys-webhook-example/internal/logger"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// GetClientSet 获取集群CURD
func GetClientSet() *kubernetes.Clientset {
	// 获取集群内配置,pod创建的时候会把sa token 挂在到容器内/var/run/secrets/kubernetes.io/serviceaccoun目录下InClusterConfig函数就是在这里去找配置
	lg := logger.WithName("GetClientSet")
	config, err := rest.InClusterConfig()
	if err != nil {
		lg.Debug("Failed to initialize InClusterConfig", zap.Error(err))
		// 如果集群内配置失败，则尝试使用本地的 kubeConfig 文件
		kubeConfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
		if err != nil {
			lg.Fatal("Failed to initialize clientSet from local kubeConfig", zap.Error(err))
			return nil // 返回错误给调用者处理
		}
	}
	// 根据配置信息创建client，client可以操作各种资源的CURD
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		lg.Fatal("Failed to initialize clientSet", zap.Error(err))
		return nil // 返回错误给调用者处理
	}
	return clientSet
}
