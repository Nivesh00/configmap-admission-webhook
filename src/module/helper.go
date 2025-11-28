package module

import (
	"encoding/json"
	"io"
    "net/http"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
)

// Function parses an admission request and returns the
// admission review, the configmap and an error
func ParseAdmissionRequest(r *http.Request) (*admissionv1.AdmissionReview, *corev1.ConfigMap, error) {

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		Logger.Error("could not parse request body")
		return nil, nil, err
	}

	Logger.Info("successfully parsed request body")

	// Assign admission review to object, note that admissionReview is the upstream
	// and e.g. pod could be used
	var admissionReview admissionv1.AdmissionReview
	if err := json.Unmarshal([]byte(body), &admissionReview); err != nil {
		Logger.Info("could unmarshall request body")
		return nil, nil, err
	}

	Logger.Info(
		"successfully unmarshalled request body",
		"name", 	 admissionReview.Request.Name,
		"namespace", admissionReview.Request.Namespace,
		"group", 	 admissionReview.Request.Kind.Group,
		"kind", 	 admissionReview.Request.Kind.Kind,
		"operation", admissionReview.Request.Operation,
	)

	// Assign admission request object to specific k8s object
	var configmap corev1.ConfigMap
	if err := json.Unmarshal(admissionReview.Request.Object.Raw, &configmap); err != nil {
		Logger.Info("could not parse k8s object")
		return nil, nil, err
	}

	return &admissionReview, &configmap, nil
}