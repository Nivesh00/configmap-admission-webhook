package main

import (
	"net/http"
	"os"
    "log/slog"
	"github.com/Nivesh00/configmap-admission-webhook.git/module"
)

func main() {

    slog.Info("Starting up program")

    // Look for forbidden keys in environmental variables
    // and create a global variable from it
    err := module.AssignForbiddenKeys()
    if err != nil {
        slog.Error("a fatal problem occured, cannot continue", slog.Any("error", err))
        os.Exit(1)
    }

    http.HandleFunc("/validate-configmap-keys", module.HandleValidation)
    http.HandleFunc("/mutate-configmap-keys", module.HandleMutation)

    slog.Info("listening on port :443")
    err = http.ListenAndServeTLS(":443", "server.crt", "server.key", nil)
    slog.Error("an error occured, stopping server", slog.Any("error", err))
    os.Exit(1)
}


