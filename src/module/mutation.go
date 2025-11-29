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

func HandleMutation(w http.ResponseWriter, r *http.Request) {

	admissionReviewOld, configmap, err := ParseAdmissionRequest(r)
	if err != nil {
		Logger.Error(
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

	Logger.Info(
		"proceeding to mutation of object",
		"name", 	 objName,
		"namespace", objNamespace,
		"group", 	 objGroup,
		"kind",  	 objKind,
	)

	// Annotations which will show removed keys
	auditAnnotations := make(map[string]string, 2)
	var keysRemoved []string

	// Allow
	allowed := true

	// Patches operations to do
	var patches []PatchOperation

	// User settings
	forbiddenKeys := &GlobalForbiddenKeys.KeyList
	caseSensitive :=  GlobalForbiddenKeys.CaseSensitive
	policy        :=  GlobalForbiddenKeys.Policy

	// Remove forbidden keys if policy is set to auto
	if policy == "AUTO" {

		for key := range(configmap.Data) {

			keyCheck := key
			// Ignore case if case sensitive is false
			if !caseSensitive {
				keyCheck = strings.ToLower(key)
			}
			// Reject if key is forbidden
			if slices.Contains(*forbiddenKeys, keyCheck) {
				Logger.Debug(
					"found forbidden key during mutation which will be removed",
					"name", 	 objName,
					"namespace", objNamespace,
					"group", 	 objGroup,
					"kind",  	 objKind,
					"key", 		 key,
				)
				// Remove path
				patchOperation := PatchOperation{
					Operation: "remove",
					Path: 	   "/data/" + key,
				}
				// Append to patches list
				patches = append(patches, patchOperation)
				// Add key to warning
				keysRemoved = append(keysRemoved, key)
			}
		}
	}

	// Add annotations to object
	auditAnnotations["policy"] 		= policy
	auditAnnotations["keysRemoved"] = strings.Join(keysRemoved, ", ")

	// Serialize patch operation
	patchesBytes, err := json.Marshal(patches)
	if err != nil {
		Logger.Error("could not serialize patch operation", slog.Any("error", err))
		allowed = false
	}

	Logger.Debug(
		"patches operations to be performed",
		slog.Any("patches", string(patchesBytes)),
	)

	patchType := admissionv1.PatchTypeJSONPatch

	// Create admission response
	admissionResponse := admissionv1.AdmissionResponse{
		UID: 		 	  admissionRequest.UID,
		Allowed: 	      allowed,
		AuditAnnotations: auditAnnotations,
		Patch: 			  patchesBytes,
		PatchType: 	 	  &patchType,
	}

	// Result for user
	if !allowed {
		result := &metav1.Status{
			Status: "Failure",
			Message: "error=" + err.Error(),
			Code: 406,
		}
		admissionResponse.Result = result
	}

	// User warnings
	if policy == "AUTO" {
		warnings := []string{
			"policy is set to 'AUTO', any forbidden key found will be removed before applying configmap",
		}
		admissionResponse.Warnings = warnings
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
		Logger.Error(
			"cannot marshal response",
			"name", 	 objName,
			"namespace", objNamespace,
			"group", 	 objGroup,
			"kind",  	 objKind,
			slog.Any("error", err),
		)
	}

	Logger.Info(
		"mutation done",
		"name", 	 objName,
		"namespace", objNamespace,
		"group", 	 objGroup,
		"kind",  	 objKind,
	)

	w.Write(responseBytes)
}