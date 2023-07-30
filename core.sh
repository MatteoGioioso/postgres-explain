#!/usr/bin/env bash

echo "// THIS FILE IS AUTO GENERATED
package main
import (
	\"postgres-explain/backend/core\"
	\"postgres-explain/backend/modules\"
)

func GetModules() (map[string]modules.Module, error) {
	return core.CoreModules, nil
}" > backend/get_modules.go
