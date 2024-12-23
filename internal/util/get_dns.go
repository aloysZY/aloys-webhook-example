package util

import (
	"context"
	"fmt"
	"sync"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	localDnsBindAddress string    // 用于存储 bind 值
	coreDNSBindAddress  string    // 用于存储 CoreDNS Service 的 IP 值
	initialized         bool      // 初始化标志
	mu                  sync.Once // 确保初始化只执行一次
)

// GetDNSIP 获取 localDns 的 bind 值和coreDNS并缓存它
func GetDNSIP() (string, string, error) {
	var err error
	mu.Do(func() {
		// 只有在第一次调用时执行以下代码
		clientSet := GetClientSet()
		localDnsBindAddress, err = getLocalIPFromDaemonSet(clientSet)
		coreDNSBindAddress, err = getCoreIPFromService(clientSet)
		if err == nil {
			initialized = true
		}
	})

	if !initialized {
		return "", "", fmt.Errorf("failed to initialize coreDNS bind value: %v", err)
	}

	return localDnsBindAddress, coreDNSBindAddress, nil
}

// getLocalIPFromDaemonSet 获取 DaemonSet 中指定容器的 -localip 参数值
func getLocalIPFromDaemonSet(clientSet *kubernetes.Clientset) (string, error) {
	// 获取指定命名空间中的 DaemonSet
	localDNSDS, err := clientSet.AppsV1().DaemonSets("kube-system").Get(context.TODO(), "node-local-dns", metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return "", fmt.Errorf("daemonset node-local-dns not found in namespace kube-system")
		}
		return "", fmt.Errorf("failed to get daemonset node-local-dns: %v", err)
	}

	// 遍历所有容器，查找包含 -localip 参数的容器
	for _, container := range localDNSDS.Spec.Template.Spec.Containers {
		args := container.Args
		for i := 0; i < len(args)-1; i++ {
			if args[i] == "-localip" {
				return args[i+1], nil
			}
		}
	}

	return "", fmt.Errorf("localip parameter not found in daemonset node-local-dns containers")
}

// getCoreIPFromService 获取coreNDs ip
func getCoreIPFromService(clientSet *kubernetes.Clientset) (string, error) {
	// 获取 CoreDNS Service
	coreDNSService, err := clientSet.CoreV1().Services("kube-system").Get(context.TODO(), "kube-dns", metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return "", fmt.Errorf("service kube-dns not found in namespace kube-system")
		}
		return "", fmt.Errorf("failed to get service kube-dns: %v", err)
	}

	// 获取 CoreDNS Service 的 Cluster IP
	return coreDNSService.Spec.ClusterIP, nil
}
