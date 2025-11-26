package module

import (
	"os"
	"strings"
	"log/slog"
    "fmt"
)

// Global variable for user settings
var GlobalForbiddenKeys ForbiddenKeys

func AssignForbiddenKeys() error {

	// Get environmental variable
    keys := os.Getenv("FORBIDDEN_KEYS")
    if keys == "" {
        slog.Info("Cannot find any forbidden keys")
        return fmt.Errorf("no forbidden keys found, cannot run program")
    }
    caseSensitive := os.Getenv("CASE_SENSITIVE")
    policy := os.Getenv("POLICY")

    // Assign forbidden keys struct
    GlobalForbiddenKeys.CreateForbiddenKeyList(policy, caseSensitive, keys)

	slog.Info(
		"Finished processing user settings",
		"forbidden keys",
		"[" + strings.Join(GlobalForbiddenKeys.KeyList, ", ") + "]",
        "policy",
        policy,
        "case sensitive",
        caseSensitive,
	)

    return nil
}


// Object to store for testing
type ForbiddenKeys struct {
    KeyList        []string
    Policy           string
    CaseSensitive    bool
}


// Looks for environmental variables and adds them to
// the key list
func (f *ForbiddenKeys) CreateForbiddenKeyList(policy string, caseSensitive string, keys string) {

    // Set policy
    f.Policy = "auto"
    policy = strings.ToLower(policy)
    if policy == "manual" {
        f.Policy = "manual"
    }

    // Set case sensitivity
    f.CaseSensitive = true
    if caseSensitive == "disabled" {
        f.CaseSensitive = false
    }

	// If case sensitive, store all values as lowercase
	if !f.CaseSensitive {
		keys = strings.ToLower(keys)
	}
    f.KeyList = strings.Split(keys, ",")
}