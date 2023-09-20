package enterprise

import (
	"postgres-explain/backend/core/info"
	"postgres-explain/backend/enterprise/collector"
	"postgres-explain/backend/enterprise/query_explainer"
	"postgres-explain/backend/modules"
)

var EnterpriseModules = map[string]modules.Module{
	query_explainer.ModuleName: &query_explainer.Module{},
	collector.ModuleName:       &collector.Module{},
	info.ModuleName:            &info.Module{},
}
