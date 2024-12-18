package setting

import (
	"k8s.io/api/admission/v1"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

var (
	scheme = runtime.NewScheme()
	Codecs = serializer.NewCodecFactory(scheme)
)

func init() {
	addToScheme(scheme)
}

// addToScheme adds all necessary API versions and resource types to the provided scheme.
func addToScheme(scheme *runtime.Scheme) {
	addToSchemeFuncs := map[string]func(*runtime.Scheme) error{
		"corev1":                       corev1.AddToScheme,
		"admissionv1beta1":             admissionv1beta1.AddToScheme,
		"admissionregistrationv1beta1": admissionregistrationv1beta1.AddToScheme,
		"admissionv1":                  v1.AddToScheme,
		"admissionregistrationv1":      admissionregistrationv1.AddToScheme,
	}

	for _, addFunc := range addToSchemeFuncs {
		utilruntime.Must(addFunc(scheme))
	}
}
