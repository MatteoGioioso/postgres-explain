package enterprise

import (
	"postgres-explain/backend/enterprise/activities"
	"postgres-explain/backend/enterprise/analytics"
	"postgres-explain/backend/enterprise/collector"
	"postgres-explain/backend/enterprise/info"
	"postgres-explain/backend/enterprise/query_explainer"
	"postgres-explain/backend/modules"
)

var EnterpriseModules = map[string]modules.Module{
	query_explainer.ModuleName: &query_explainer.Module{},
	collector.ModuleName:       &collector.Module{},
	info.ModuleName:            &info.Module{},
	activities.ModuleName:      &activities.Module{},
	analytics.ModuleName:       &analytics.Module{},
}
