package main

import (
	"log/slog"
	"net/http"
	"os"
	"github.com/Nivesh00/configmap-admission-webhook/src/module"
)

func main() {

    slog.Info("starting up program")

    // Look for forbidden keys in environmental variables
    // and create a global variable from it
    err := module.AssignForbiddenKeys()
    if err != nil {
        slog.Error("a fatal problem occured, cannot continue", slog.Any("error", err))
        os.Exit(1)
    }

    http.HandleFunc("/validate", module.HandleValidation)
    http.HandleFunc("/mutate", module.HandleMutation)

    // TLS server
    slog.Info("listening on port :443")
    err = http.ListenAndServeTLS(
        ":443", 
        "/etc/certs/tls.crt", 
        "/etc/certs/tls.key", 
        nil,
    )
    slog.Error("an error occured, stopping server", slog.Any("error", err))
    os.Exit(1)
}

