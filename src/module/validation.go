package module

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	admissionv1 "k8s.io/api/admission/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func HandleValidation(w http.ResponseWriter, r *http.Request) {

	admissionReviewOld, configmap, err := ParseAdmissionRequest(r)
	if err != nil {
		slog.Error(
			"an error occured, cannot validate object",
			"name", 	 configmap.GetName(),
			"namespace", configmap.GetNamespace(),
			"group", 	 configmap.GetObjectKind().GroupVersionKind().Group,
			"kind", 	 configmap.GetObjectKind().GroupVersionKind().Kind,
			slog.Any("error", err),
		)
		return
	}
	admissionRequest := admissionReviewOld.Request
	
	// Save logging info
	objName  	 := configmap.GetName()
	objNamespace := configmap.GetNamespace()
	objGroup 	 := configmap.GetObjectKind().GroupVersionKind().Group
	objKind  	 := configmap.GetObjectKind().GroupVersionKind().Kind

	slog.Info(
		"proceeding to validation of object",
		"name", 	 objName,
		"namespace", objNamespace,
		"group", 	 objGroup,
		"kind",  	 objKind,
	)

	// Variable to check if operation is allowed
	allowed := true

	// Result to give back to user
	var forbiddenKeysFound []string

	// User settings
	forbiddenKeys := &GlobalForbiddenKeys.KeyList
	caseSensitive :=  GlobalForbiddenKeys.CaseSensitive
	policy 		  :=  GlobalForbiddenKeys.Policy

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
				"name", 	 objName,
				"namespace", objNamespace,
				"group", 	 objGroup,
				"kind",  	 objKind,
				"key", 		 key,
			)
			// Append warning
			forbiddenKeysFound = append(forbiddenKeysFound, key)
			allowed = false
		}
	}

	// Result for user in case object is invalid
	caseSetting := "on"
	if !caseSensitive {
		caseSetting = "off"
	}
	result := &metav1.Status{
		Status: "failure",
		Message: 
			"Using policy " + policy + " and case sensitive " + caseSetting +
			", forbidden keys were found during validation: [" +
			strings.Join(forbiddenKeysFound, ", ") + "]",
		Code: 406,
	}

	// Create admission response
	admissionResponse := admissionv1.AdmissionResponse{
		UID: 	  admissionRequest.UID,
		Allowed:  allowed,
		Result:   result,
	}

	// Create admission review response
	admissionReviewNew := admissionv1.AdmissionReview{
		Response: &admissionResponse,
	}

	// Set group version
	admissionReviewNew.SetGroupVersionKind(admissionReviewOld.GroupVersionKind())

	// Convert response to bytes
	responseBytes, err := json.Marshal(&admissionReviewNew)
	if err != nil {
		slog.Error(
			"cannot marshal response",
			"name", 	 objName,
			"namespace", objNamespace,
			"group", 	 objGroup,
			"kind",  	 objKind,
			slog.Any("error", err),
		)
	}

	slog.Info(
		"validation done",
		"name", 	 objName,
		"namespace", objNamespace,
		"group", 	 objGroup,
		"kind",  	 objKind,
	)

	w.Write(responseBytes)
}