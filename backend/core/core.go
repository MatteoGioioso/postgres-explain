package core

import (
	"postgres-explain/backend/core/analytics"
	"postgres-explain/backend/core/info"
	"postgres-explain/backend/core/query_explainer"
	"postgres-explain/backend/modules"
)

var CoreModules = map[string]modules.Module{
	query_explainer.ModuleName: &query_explainer.Module{},
	analytics.ModuleName:       &analytics.Module{},
	info.ModuleName:            &info.Module{},
}
