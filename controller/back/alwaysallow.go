/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package back

import (
	"fmt"
	"time"

	"github.com/aloys.zy/aloys-webhook-example/internal/setting"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

// AlwaysAllowDelayFiveSeconds sleeps for five seconds and allows all requests made to this function.
func AlwaysAllowDelayFiveSeconds(ar admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	nodeResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "nodes"}
	if ar.Request.Resource != nodeResource {
		err := fmt.Errorf("expect resource to be %s", nodeResource)
		klog.Error(err)
		return setting.ToV1AdmissionResponse(err)
	}
	klog.V(2).Info("always-allow-with-delay sleeping for 5 seconds")
	time.Sleep(5 * time.Second)
	klog.V(2).Info("calling always-allow")
	reviewResponse := admissionv1.AdmissionResponse{
		Allowed: true,
		Result:  &metav1.Status{Message: "this webhook-template allows all requests"},
	}
	return &reviewResponse
}
