package handlefunc

import (
	"net/http"

	"github.com/aloys.zy/aloys-webhook-example/internal/controller"
)

// ServeAlwaysAllowDelayFiveSeconds 传入请求参数
func ServeAlwaysAllowDelayFiveSeconds(w http.ResponseWriter, r *http.Request) {
	serve(w, r, newDelegateToV1AdmitHandler(controller.AlwaysAllowDelayFiveSeconds))
}

func ServeAlwaysDeny(w http.ResponseWriter, r *http.Request) {
	serve(w, r, newDelegateToV1AdmitHandler(controller.AlwaysDeny))
}

func ServeAddLabel(w http.ResponseWriter, r *http.Request) {
	serve(w, r, newDelegateToV1AdmitHandler(controller.AddLabel))
}

func ServePods(w http.ResponseWriter, r *http.Request) {
	serve(w, r, newDelegateToV1AdmitHandler(controller.AdmitPods))
}

func ServeAttachingPods(w http.ResponseWriter, r *http.Request) {
	serve(w, r, newDelegateToV1AdmitHandler(controller.DenySpecificAttachment))
}

func ServeMutatePods(w http.ResponseWriter, r *http.Request) {
	serve(w, r, newDelegateToV1AdmitHandler(controller.MutatePods))
}

func ServeMutatePodsSidecar(w http.ResponseWriter, r *http.Request) {
	serve(w, r, newDelegateToV1AdmitHandler(controller.MutatePodsSidecar))
}

func ServeConfigmaps(w http.ResponseWriter, r *http.Request) {
	serve(w, r, newDelegateToV1AdmitHandler(controller.AdmitConfigMaps))
}

func ServeMutateConfigmaps(w http.ResponseWriter, r *http.Request) {
	serve(w, r, newDelegateToV1AdmitHandler(controller.MutateConfigmaps))
}

func ServeCustomResource(w http.ResponseWriter, r *http.Request) {
	serve(w, r, newDelegateToV1AdmitHandler(controller.AdmitCustomResource))
}

func ServeMutateCustomResource(w http.ResponseWriter, r *http.Request) {
	serve(w, r, newDelegateToV1AdmitHandler(controller.MutateCustomResource))
}

func ServeCRD(w http.ResponseWriter, r *http.Request) {
	serve(w, r, newDelegateToV1AdmitHandler(controller.AdmitCRD))
}

func ServeMutateNodeOversold(writer http.ResponseWriter, request *http.Request) {
	serve(writer, request, newDelegateToV1AdmitHandler(controller.MutateNodes))
}

func ServeValidatePodContainerLimit(writer http.ResponseWriter, request *http.Request) {
	serve(writer, request, newDelegateToV1AdmitHandler(controller.ValidatePodContainerLimit))
}
