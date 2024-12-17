/*
Copyright 2018 The Kubernetes Authors.

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
	"github.com/aloys.zy/aloys-webhook-example/api"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

const (
	configMapPatch1 string = `[
         { "op": "add", "path": "/data/mutation-stage-1", "value": "yes" }
     ]`
	configMapPatch2 string = `[
         { "op": "add", "path": "/data/mutation-stage-2", "value": "yes" }
     ]`
)

// deny configmaps with specific key-value pair.
func AdmitConfigMaps(ar admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	klog.V(2).Info("admitting configmaps")
	configMapResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
	if ar.Request.Resource != configMapResource {
		klog.Errorf("expect resource to be %s", configMapResource)
		return nil
	}

	var raw []byte
	if ar.Request.Operation == admissionv1.Delete {
		raw = ar.Request.OldObject.Raw
	} else {
		raw = ar.Request.Object.Raw
	}
	configmap := corev1.ConfigMap{}
	deserializer := api.Codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(raw, nil, &configmap); err != nil {
		klog.Error(err)
		return api.ToV1AdmissionResponse(err)
	}
	reviewResponse := admissionv1.AdmissionResponse{}
	reviewResponse.Allowed = true
	for k, v := range configmap.Data {
		if k == "webhook-e2e-test" && v == "webhook-disallow" &&
			(ar.Request.Operation == admissionv1.Create || ar.Request.Operation == admissionv1.Update) {
			reviewResponse.Allowed = false
			reviewResponse.Result = &metav1.Status{
				Reason: "the configmap contains unwanted key and value",
			}
		}
		if k == "webhook-e2e-test" && v == "webhook-nondeletable" && ar.Request.Operation == admissionv1.Delete {
			reviewResponse.Allowed = false
			reviewResponse.Result = &metav1.Status{
				Reason: "the configmap cannot be deleted because it contains unwanted key and value",
			}
		}
	}
	return &reviewResponse
}

func MutateConfigmaps(ar admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	klog.V(2).Info("mutating configmaps")
	configMapResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
	if ar.Request.Resource != configMapResource {
		klog.Errorf("expect resource to be %s", configMapResource)
		return nil
	}

	raw := ar.Request.Object.Raw
	configmap := corev1.ConfigMap{}
	deserializer := api.Codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(raw, nil, &configmap); err != nil {
		klog.Error(err)
		return api.ToV1AdmissionResponse(err)
	}
	reviewResponse := admissionv1.AdmissionResponse{}
	reviewResponse.Allowed = true
	if configmap.Data["mutation-start"] == "yes" {
		reviewResponse.Patch = []byte(configMapPatch1)
	}
	if configmap.Data["mutation-stage-1"] == "yes" {
		reviewResponse.Patch = []byte(configMapPatch2)
	}

	if len(reviewResponse.Patch) != 0 {
		pt := admissionv1.PatchTypeJSONPatch
		reviewResponse.PatchType = &pt
	}

	return &reviewResponse
}
