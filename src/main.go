package main

import (
	"log/slog"
	"net/http"
    "fmt"
	"os"
    "flag"
	"github.com/Nivesh00/configmap-manager/src/module"
)

func main() {
    
    fmt.Println("starting up server...")

    // Flags
    logLevel := flag.String("log-level", "warn", "log level, default is warn")
    port     := flag.String("port", "443", "port server listens to")
    flag.Parse()
    module.CreateLogger(logLevel)

    // Look for forbidden keys in environmental variables
    // and create a global variable from it
    err := module.AssignForbiddenKeys()
    if err != nil {
        module.Logger.Error("a fatal problem occured, cannot continue", slog.Any("error", err))
        os.Exit(1)
    }

    http.HandleFunc("/validate", module.HandleValidation)
    http.HandleFunc("/mutate", module.HandleMutation)

    // TLS server
    fmt.Println("server now listening on port " + *port)
    err = http.ListenAndServeTLS(
        ":" + *port, 
        "/etc/certs/tls.crt", 
        "/etc/certs/tls.key", 
        nil,
    )
    module.Logger.Error("an error occured, server shutting down", slog.Any("error", err))
    os.Exit(1)
}

