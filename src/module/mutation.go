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
		"proceeding to mutation of object",
		"name", 	 objName,
		"namespace", objNamespace,
		"group", 	 objGroup,
		"kind",  	 objKind,
	)

	// Annotations which will show removed keys
	auditAnnotations := make(map[string]string, 2)
	var keysRemoved []string

	// Warnings to give back to user
	var forbiddenKeysFound []string

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
				slog.Info(
					"found forbidden key during mutation which will be removed",
					"name", 	 objName,
					"namespace", objNamespace,
					"group", 	 objGroup,
					"kind",  	 objKind,
					"key", 		 key,
				)
				// Append warning
				forbiddenKeysFound = append(forbiddenKeysFound, key)
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

	// Warning for user
	msg := 
		"Forbidden keys found and removed during mutation: [" +
	 	strings.Join(forbiddenKeysFound, ", ") +
		"]"

	patchType := admissionv1.PatchTypeJSONPatch

	// Create admission response
	admissionResponse := admissionv1.AdmissionResponse{
		UID: 		 	  admissionRequest.UID,
		Allowed: 	      true,
		Warnings: 		  []string{msg},
		AuditAnnotations: auditAnnotations,
		Patch: 			  patchesBytes,
		PatchType: 	 	  &patchType,
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
		"mutation done",
		"name", 	 objName,
		"namespace", objNamespace,
		"group", 	 objGroup,
		"kind",  	 objKind,
	)

	w.Write(responseBytes)
}