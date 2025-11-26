package module

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	admissionv1 "k8s.io/api/admission/v1"
)

func HandleValidation(w http.ResponseWriter, r *http.Request) {

	admissionReview, configmap, err := ParseAdmissionRequest(r)
	if err != nil {
		slog.Error(
			"an error occured, cannot validate object",
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

	slog.Info(
		"proceeding to validation of object",
		"name",
		configmap.GetName(),
		"namespace",
		configmap.GetNamespace(),
		"kind",
		configmap.GetObjectKind(),
	)

	// Variable to check if operation is allowed
	allowed := true

	// Warnings to give back to user
	var warnings []string

	// User settings
	forbiddenKeys := &GlobalForbiddenKeys.KeyList
	caseSensitive :=  GlobalForbiddenKeys.CaseSensitive
	policy        :=  GlobalForbiddenKeys.Policy

	// Check for forbidden keys
	for key := range(configmap.Data) {

		keyCheck := key
		// Ignore case if case sensitive is false
		if !caseSensitive {
			keyCheck = strings.ToLower(key)
		}
		// Reject if key is forbidden
		if slices.Contains(*forbiddenKeys, keyCheck) {
			slog.Info(
				"found forbidden key during validation",
				"name",
				configmap.GetName(),
				"namespace",
				configmap.GetNamespace(),
				"kind",
				configmap.GetObjectKind(),
				"key",
				key,
			)
			allowed = false
			warnings = append(warnings, key)
		}
	}

	// If object is rejected
	if !allowed {
		msg := "forbidden keys found in configmap for policy " + policy
		// Prepend msg to warnings
		warnings = append(
			[]string{msg}, 
			warnings...,
		)
	}

	// Create admission response
	admissionResponse := admissionv1.AdmissionResponse{
		UID: admissionReview.Request.UID,
		Allowed: allowed,
		Warnings: warnings,
	}

	// Create admission review response
	admissionReviewResponse := admissionv1.AdmissionReview{
		Response: &admissionResponse,
	}

	// Convert response to bytes
	responseBytes, err := json.Marshal(&admissionReviewResponse)
	if err != nil {
		slog.Error(
			"cannot marshal response",
			"name",
			configmap.GetName(),
			"namespace",
			configmap.GetNamespace(),
			"kind",
			configmap.GetObjectKind(),
			slog.Any("error", err),
		)
	}

	w.Write(responseBytes)
}