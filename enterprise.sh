#!/usr/bin/env bash

echo "// THIS FILE IS AUTO GENERATED
package main
import (
	\"postgres-explain/backend/enterprise\"
	\"postgres-explain/backend/modules\"
)

func GetModules() (map[string]modules.Module, error) {
	return enterprise.EnterpriseModules, nil
}" > backend/get_modules.go
