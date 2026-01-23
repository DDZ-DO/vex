#!/bin/bash

# Run-Script für vex
# Verwendung: ./scripts/run.sh [datei]

cd "$(dirname "$0")/.."

# Bauen und starten
go run ./cmd/vex "$@"
