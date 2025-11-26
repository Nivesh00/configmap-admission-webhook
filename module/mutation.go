package module

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	admissionv1 "k8s.io/api/admission/v1"
)

func HandleMutation(w http.ResponseWriter, r *http.Request) {

	admissionReview, configmap, err := ParseAdmissionRequest(r)
	if err != nil {
		slog.Error(
			"An error occured, cannot validate object",
			"name",
			configmap.GetName(),
			"namespace",
			configmap.GetNamespace(),
			"kind",
			configmap.GetObjectKind(),
			slog.Any("error", err),
		)
		return
	}

	// Annotations which will show removed keys
	auditAnnotations := make(map[string]string, 2)
	var keysRemoved []string

	// Patches operations to do
	var patches []string

	// User settings
	forbiddenKeys := &GlobalForbiddenKeys.KeyList
	caseSensitive :=  GlobalForbiddenKeys.CaseSensitive
	policy        :=  GlobalForbiddenKeys.Policy

	// Remove forbidden keys if policy is set to auto
	if policy == "auto" {

		for key := range(configmap.Data) {

			keyCheck := key
			// Ignore case if case sensitive is false
			if !caseSensitive {
				keyCheck = strings.ToLower(key)
			}
			// Reject if key is forbidden
			if slices.Contains(*forbiddenKeys, keyCheck) {
				// Remove path
				patchOperation := "{'op': 'remove', 'path': '/spec/data/" + key + "'}"
				// Append to patches slice
				patches = append(patches, patchOperation)
				// Add key to warning
				keysRemoved = append(keysRemoved, key)
			}
		}
	}

	// Add annotations to object
	auditAnnotations["policy"] 		= policy
	auditAnnotations["keysRemoved"] = strings.Join(keysRemoved, ", ")
	// Convert patches string then to bytes
	patchesUnicode := "[" + strings.Join(patches, ",") + "]"
	patchesBytes := []byte(patchesUnicode)

	// Create admission response
	admissionResponse := admissionv1.AdmissionResponse{
		UID: admissionReview.Request.UID,
		Allowed: true,
		AuditAnnotations: auditAnnotations,
		Patch: patchesBytes,
	}

	// Create admission review response
	admissionReviewResponse := admissionv1.AdmissionReview{
		Response: &admissionResponse,
	}

	// Convert response to bytes
	responseBytes, err := json.Marshal(&admissionReviewResponse)
	if err != nil {
		slog.Error("Cannot marshal response", slog.Any("error", err))
	}

	w.Write(responseBytes)
}