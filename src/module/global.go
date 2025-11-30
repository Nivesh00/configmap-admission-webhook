package module

import (
	"os"
	"strings"
	"log/slog"
    "fmt"
)

// Global logger
var Logger *slog.Logger
// Global variable for user settings
var GlobalForbiddenKeys ForbiddenKeys

// Create logger
func CreateLogger(lvl *string) {

    logLevel := new(slog.LevelVar)
    // Set log level
    switch *lvl {
    case "debug":
        logLevel.Set(slog.LevelDebug)
    case "info":
        logLevel.Set(slog.LevelInfo)
    case "error":
        logLevel.Set(slog.LevelError)
    // Default is warn
    default:
        logLevel.Set(slog.LevelWarn)
    }

    // Create logger
    Logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
        Level: logLevel,
    }))

    Logger.Debug("successfully created logger")
}

func AssignForbiddenKeys() error {

	// Get environmental variable
    keys := os.Getenv("FORBIDDEN_KEYS")
    if keys == "" {
        Logger.Info("Cannot find any forbidden keys")
        return fmt.Errorf("no forbidden keys found, cannot run program")
    }
    caseSensitive := os.Getenv("CASE_SENSITIVE")
    policy := os.Getenv("POLICY")

    // Assign forbidden keys struct
    GlobalForbiddenKeys.CreateForbiddenKeyList(policy, caseSensitive, keys)

	Logger.Info(
		"Finished processing user settings",
		"FORBIDDEN_KEYS",
		"[" + strings.Join(GlobalForbiddenKeys.KeyList, ", ") + "]",
        "POLICY",
        GlobalForbiddenKeys.Policy,
        "CASE_SENSITIVE",
        GlobalForbiddenKeys.CaseSensitive,
	)

    return nil
}

// Object to store for testing
type ForbiddenKeys struct {
    KeyList        []string
    Policy           string
    CaseSensitive    bool
}

// Patches operations
type PatchOperation struct {
    Operation string `json:"op"`
    Path      string `json:"path"`
    Value     string `json:"value,omitempty"`
}

// Looks for environmental variables and adds them to
// the key list
func (f *ForbiddenKeys) CreateForbiddenKeyList(policy string, caseSensitive string, keys string) {

    // Set policy
    f.Policy = "MANUAL"
    policy = strings.ToUpper(policy)
    if policy == "AUTO" {
        f.Policy = "AUTO"
    }

    // Set case sensitivity
    f.CaseSensitive = true
    caseSensitive = strings.ToUpper(caseSensitive)
    if caseSensitive == "FALSE" {
        f.CaseSensitive = false
    }

	// If case sensitive, store all values as lowercase
    keys = strings.ReplaceAll(keys, " ", "")
	if !f.CaseSensitive {
		keys = strings.ToLower(keys)
	}
    f.KeyList = strings.Split(keys, ",")
}

