#!/bin/bash

# Debug-Script für vex
# Verwendung: ./scripts/debug.sh [datei]

cd "$(dirname "$0")/.."

# Go bin zum PATH hinzufügen
export PATH="$PATH:$(go env GOPATH)/bin"

# Prüfen ob Delve installiert ist
if ! command -v dlv &> /dev/null; then
    echo "Delve nicht gefunden. Installiere..."
    go install github.com/go-delve/delve/cmd/dlv@latest
fi

# Debug starten
if [ -n "$1" ]; then
    echo "Starte Debug mit Datei: $1"
    dlv debug ./cmd/vex -- "$1"
else
    echo "Starte Debug ohne Datei"
    dlv debug ./cmd/vex
fi
