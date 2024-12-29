package setting

import (
	admissionv1 "k8s.io/api/admission/v1"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConvertAdmissionRequestToV1 converts an admission v1beta1.AdmissionRequest to v1.AdmissionRequest.
func ConvertAdmissionRequestToV1(r *admissionv1beta1.AdmissionRequest) *admissionv1.AdmissionRequest {
	if r == nil {
		return nil
	}
	v1Req := &admissionv1.AdmissionRequest{
		Kind:               r.Kind,
		Namespace:          r.Namespace,
		Name:               r.Name,
		Object:             r.Object,
		Resource:           r.Resource,
		Operation:          admissionv1.Operation(r.Operation),
		UID:                r.UID,
		DryRun:             r.DryRun,
		OldObject:          r.OldObject,
		Options:            r.Options,
		RequestKind:        r.RequestKind,
		RequestResource:    r.RequestResource,
		RequestSubResource: r.RequestSubResource,
		SubResource:        r.SubResource,
		UserInfo:           r.UserInfo,
	}
	return v1Req
}

// ConvertAdmissionResponseToV1beta1 converts an admission v1.AdmissionResponse to v1beta1.AdmissionResponse.
func ConvertAdmissionResponseToV1beta1(r *admissionv1.AdmissionResponse) *admissionv1beta1.AdmissionResponse {
	if r == nil {
		return nil
	}
	var pt *admissionv1beta1.PatchType
	if r.PatchType != nil {
		t := admissionv1beta1.PatchType(*r.PatchType)
		pt = &t
	}

	v1beta1Resp := &admissionv1beta1.AdmissionResponse{
		UID:              r.UID,
		Allowed:          r.Allowed,
		AuditAnnotations: r.AuditAnnotations,
		Patch:            r.Patch,
		PatchType:        pt,
		Result:           r.Result,
		Warnings:         r.Warnings,
	}

	return v1beta1Resp
}

// ToV1AdmissionResponse creates a v1.AdmissionResponse from an error.
func ToV1AdmissionResponse(err error) *admissionv1.AdmissionResponse {
	if err == nil {
		return &admissionv1.AdmissionResponse{Allowed: true}
	}
	return &admissionv1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}
