package api

import (
	"net/http"

	"github.com/aloys.zy/aloys-webhook-example/internal/routers"
	ctrl "sigs.k8s.io/controller-runtime"
)

// 辅助函数：根据名称获取对应的处理函数
func getHandlerFuncByName(name string) http.HandlerFunc {
	switch name {
	case "ServeMutateCPUOversell":
		return routers.ServeMutateCPUOversell
	case "MutatePodDNSConfig":
		return routers.MutatePodDNSConfig
	case "ServeAlwaysAllowDelayFiveSeconds":
		return routers.ServeAlwaysAllowDelayFiveSeconds
	case "ServeAlwaysDeny":
		return routers.ServeAlwaysDeny
	case "ServeAddLabel":
		return routers.ServeAddLabel
	case "ServePods":
		return routers.ServePods
	case "ServeAttachingPods":
		return routers.ServeAttachingPods
	case "ServeMutatePods":
		return routers.ServeMutatePods
	case "ServeMutatePodsSidecar":
		return routers.ServeMutatePodsSidecar
	case "ServeConfigmaps":
		return routers.ServeConfigmaps
	case "ServeMutateConfigmaps":
		return routers.ServeMutateConfigmaps
	case "ServeCustomResource":
		return routers.ServeCustomResource
	case "ServeMutateCustomResource":
		return routers.ServeMutateCustomResource
	case "ServeCRD":
		return routers.ServeCRD
	// case "ServeValidatePodContainerLimit":
	// 	return handlefunc.ServeValidatePodContainerLimit
	default:
		ctrl.Log.WithName("webhook Start").V(1).Info("Unknown handler name", "name", name)
		return nil
	}
}
