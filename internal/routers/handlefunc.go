package routers

import (
	"net/http"

	"github.com/aloys.zy/aloys-webhook-example/internal/controller"
	"github.com/aloys.zy/aloys-webhook-example/internal/controller/back"
	"github.com/aloys.zy/aloys-webhook-example/internal/setting"
)

// ServeMutateCPUOversell 修改Node cpu
func ServeMutateCPUOversell(writer http.ResponseWriter, request *http.Request) {
	serve(writer, request, setting.NewDelegateToV1AdmitHandler(controller.MutateCPUOversell))
}

// ServeAlwaysAllowDelayFiveSeconds 传入请求参数
func ServeAlwaysAllowDelayFiveSeconds(w http.ResponseWriter, r *http.Request) {
	serve(w, r, setting.NewDelegateToV1AdmitHandler(back.AlwaysAllowDelayFiveSeconds))
}

func ServeAlwaysDeny(w http.ResponseWriter, r *http.Request) {
	serve(w, r, setting.NewDelegateToV1AdmitHandler(back.AlwaysDeny))
}

func ServeAddLabel(w http.ResponseWriter, r *http.Request) {
	serve(w, r, setting.NewDelegateToV1AdmitHandler(back.AddLabel))
}

func ServePods(w http.ResponseWriter, r *http.Request) {
	serve(w, r, setting.NewDelegateToV1AdmitHandler(back.AdmitPods))
}

func ServeAttachingPods(w http.ResponseWriter, r *http.Request) {
	serve(w, r, setting.NewDelegateToV1AdmitHandler(back.DenySpecificAttachment))
}

func ServeMutatePods(w http.ResponseWriter, r *http.Request) {
	serve(w, r, setting.NewDelegateToV1AdmitHandler(back.MutatePods))
}

func ServeMutatePodsSidecar(w http.ResponseWriter, r *http.Request) {
	serve(w, r, setting.NewDelegateToV1AdmitHandler(back.MutatePodsSidecar))
}

func ServeConfigmaps(w http.ResponseWriter, r *http.Request) {
	serve(w, r, setting.NewDelegateToV1AdmitHandler(back.AdmitConfigMaps))
}

func ServeMutateConfigmaps(w http.ResponseWriter, r *http.Request) {
	serve(w, r, setting.NewDelegateToV1AdmitHandler(back.MutateConfigmaps))
}

func ServeCustomResource(w http.ResponseWriter, r *http.Request) {
	serve(w, r, setting.NewDelegateToV1AdmitHandler(back.AdmitCustomResource))
}

func ServeMutateCustomResource(w http.ResponseWriter, r *http.Request) {
	serve(w, r, setting.NewDelegateToV1AdmitHandler(back.MutateCustomResource))
}

func ServeCRD(w http.ResponseWriter, r *http.Request) {
	serve(w, r, setting.NewDelegateToV1AdmitHandler(back.AdmitCRD))
}

// func ServeValidatePodContainerLimit(writer http.ResponseWriter, request *http.Request) {
// 	serve(writer, request, newDelegateToV1AdmitHandler(ValidatePodContainerLimit))
// }
