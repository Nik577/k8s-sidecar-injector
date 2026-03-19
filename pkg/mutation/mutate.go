package mutation

import (
	"encoding/json"
	"log/slog"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PatchOperation represents a JSON patch operation.
type PatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// SidecarConfig allows configuring the sidecar injection.
type SidecarConfig struct {
	Image string
	Name  string
	Args  []string
}

// MutatePod handles the mutation logic for a Pod.
func MutatePod(ar *admissionv1.AdmissionReview, config SidecarConfig) *admissionv1.AdmissionResponse {
	req := ar.Request
	var pod corev1.Pod
	if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
		slog.Error("Could not unmarshal pod object", "error", err, "uid", req.UID)
		return &admissionv1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	slog.Info("AdmissionReview",
		"kind", req.Kind,
		"namespace", req.Namespace,
		"name", pod.Name,
		"uid", req.UID,
		"operation", req.Operation,
		"user", req.UserInfo.Username,
	)

	// Check if the sidecar is already injected to avoid duplicates
	for _, container := range pod.Spec.Containers {
		if container.Name == config.Name {
			slog.Info("Sidecar already injected, skipping", "pod", pod.Name, "namespace", req.Namespace)
			return &admissionv1.AdmissionResponse{
				Allowed: true,
			}
		}
	}

	// Sidecar container definition
	sidecar := corev1.Container{
		Name:  config.Name,
		Image: config.Image,
		Args:  config.Args,
		SecurityContext: &corev1.SecurityContext{
			Privileged: ptr(true),
		},
	}

	// Create JSON patch
	var patch []PatchOperation
	
	// If the containers list is empty (rare for a pod), initialize it
	// Otherwise, append to the containers list
	path := "/spec/containers/-"
	if len(pod.Spec.Containers) == 0 {
		path = "/spec/containers"
		patch = append(patch, PatchOperation{
			Op:    "add",
			Path:  path,
			Value: []corev1.Container{sidecar},
		})
	} else {
		patch = append(patch, PatchOperation{
			Op:    "add",
			Path:  path,
			Value: sidecar,
		})
	}

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		slog.Error("Could not marshal patch", "error", err, "uid", req.UID)
		return &admissionv1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	return &admissionv1.AdmissionResponse{
		Allowed: true,
		Patch:   patchBytes,
		PatchType: func() *admissionv1.PatchType {
			pt := admissionv1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

func ptr[T any](v T) *T {
	return &v
}
