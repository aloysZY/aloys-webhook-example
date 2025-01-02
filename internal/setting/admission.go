package setting

import (
	"github.com/aloys.zy/aloys-webhook-example/logger"
	"go.uber.org/zap"
	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1" // 确保导入 metav1 包
)

// admitv1beta1Func handles a v1beta1 admission review.
type admitv1beta1Func func(v1beta1.AdmissionReview) *v1beta1.AdmissionResponse

// admitv1Func handles a v1 admission review.
type admitv1Func func(admissionv1.AdmissionReview) *admissionv1.AdmissionResponse

// AdmitHandler is a handler, for both validators and mutators, that supports multiple admission review versions.
type AdmitHandler struct {
	V1beta1 admitv1beta1Func
	V1      admitv1Func
}

// NewDelegateToV1AdmitHandler creates a new AdmitHandler that delegates v1beta1 requests to the provided v1 handler function.
func NewDelegateToV1AdmitHandler(f admitv1Func) AdmitHandler {
	return AdmitHandler{
		V1beta1: delegateV1beta1AdmitToV1(f),
		V1:      f,
	}
}

// delegateV1beta1AdmitToV1 converts a v1beta1 AdmissionReview to v1, processes it with the provided v1 handler,
// and then converts the response back to v1beta1.
func delegateV1beta1AdmitToV1(f admitv1Func) admitv1beta1Func {
	lg := logger.WithName("handlefunc.delegateV1beta1AdmitToV1")

	return func(review v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
		lg.Debug("Received v1beta1 AdmissionReview", zap.Any("UID", review.Request.UID))

		// Convert v1beta1 request to v1
		v1Req := ConvertAdmissionRequestToV1(review.Request)
		if v1Req == nil {
			lg.Error("Failed to convert v1beta1 AdmissionRequest to v1")
			return &v1beta1.AdmissionResponse{
				UID:     review.Request.UID,
				Allowed: false,
				Result: &metav1.Status{
					Message: "Failed to convert v1beta1 AdmissionRequest to v1",
				},
			}
		}

		// Process the v1 request
		v1Resp := f(admissionv1.AdmissionReview{Request: v1Req})
		if v1Resp == nil {
			lg.Error("v1 handler returned a nil AdmissionResponse")
			return &v1beta1.AdmissionResponse{
				UID:     review.Request.UID,
				Allowed: false,
				Result: &metav1.Status{
					Message: "v1 handler returned a nil AdmissionResponse",
				},
			}
		}

		// Convert v1 response back to v1beta1
		v1beta1Resp := ConvertAdmissionResponseToV1beta1(v1Resp)
		if v1beta1Resp == nil {
			lg.Error("Failed to convert v1 AdmissionResponse to v1beta1")
			return &v1beta1.AdmissionResponse{
				UID:     review.Request.UID,
				Allowed: false,
				Result: &metav1.Status{
					Message: "Failed to convert v1 AdmissionResponse to v1beta1",
				},
			}
		}

		lg.Debug("Processed v1beta1 AdmissionReview successfully", zap.Any("UID", review.Request.UID))
		return v1beta1Resp
	}
}
