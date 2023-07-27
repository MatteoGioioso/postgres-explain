package query_explainer

import (
	"encoding/json"
	"math"
)

const (
	// plan property keys
	NODE_TYPE_PROP           = "Node Type"
	ACTUAL_ROWS_PROP         = "Actual Rows"
	PLAN_ROWS_PROP           = "Plan Rows"
	ACTUAL_TOTAL_TIME_PROP   = "Actual Total Time"
	ACTUAL_LOOPS_PROP        = "Actual Loops"
	TOTAL_COST_PROP          = "Total Cost"
	PLANS_PROP               = "Plans"
	STARTUP_COST             = "Startup Cost"
	PLAN_WIDTH               = "Plan Width"
	ACTUAL_STARTUP_TIME_PROP = "Actual Startup Time"
	RELATION_NAME_PROP       = "Relation Name"
	SCHEMA_PROP              = "Schema"
	ALIAS_PROP               = "Alias"
	GROUP_KEY_PROP           = "Group Key"
	SORT_KEY_PROP            = "Sort Key"
	JOIN_TYPE_PROP           = "Join Type"
	INDEX_NAME_PROP          = "Index Name"
	HASH_CONDITION_PROP      = "Hash Cond"
	EXECUTION_TIME_PROP      = "Execution Time"

	// computed
	COMPUTED_TAGS_PROP = "*Tags"

	COSTLIEST_NODE_PROP = "*Costliest Node (by cost)"
	LARGEST_NODE_PROP   = "*Largest Node (by rows)"
	SLOWEST_NODE_PROP   = "*Slowest Node (by duration)"

	MAXIMUM_COSTS_PROP         = "*Most Expensive Node (cost)"
	MAXIMUM_ROWS_PROP          = "*Largest Node (rows)"
	MAXIMUM_DURATION_PROP      = "*Slowest Node (time)"
	ACTUAL_DURATION_PROP       = "*Actual Duration"
	ACTUAL_COST_PROP           = "*Actual Cost"
	PLANNER_ESTIMATE_FACTOR    = "*Planner Row Estimate Factor"
	PLANNER_ESTIMATE_DIRECTION = "*Planner Row Estimate Direction"

	CTE_SCAN_PROP = "CTE Scan"
	CTE_NAME_PROP = "CTE Name"

	ARRAY_INDEX_KEY = "arrayIndex"

	EstimateDirectionOver  = "over"
	EstimateDirectionUnder = "under"
)

type PlanProps struct {
	NODE_TYPE_PROP           string
	ACTUAL_ROWS_PROP         string
	PLAN_ROWS_PROP           string
	ACTUAL_TOTAL_TIME_PROP   string
	ACTUAL_LOOPS_PROP        string
	TOTAL_COST_PROP          string
	PLANS_PROP               string
	STARTUP_COST             string
	PLAN_WIDTH               string
	ACTUAL_STARTUP_TIME_PROP string
	RELATION_NAME_PROP       string
	SCHEMA_PROP              string
	ALIAS_PROP               string
	GROUP_KEY_PROP           string
	SORT_KEY_PROP            string
	JOIN_TYPE_PROP           string
	INDEX_NAME_PROP          string
	HASH_CONDITION_PROP      string
	EXECUTION_TIME_PROP      string

	// computed
	COMPUTED_TAGS_PROP string

	COSTLIEST_NODE_PROP string
	LARGEST_NODE_PROP   string
	SLOWEST_NODE_PROP   string

	MAXIMUM_COSTS_PROP         string
	MAXIMUM_ROWS_PROP          string
	MAXIMUM_DURATION_PROP      string
	ACTUAL_DURATION_PROP       string
	ACTUAL_COST_PROP           string
	PLANNER_ESTIMATE_FACTOR    string
	PLANNER_ESTIMATE_DIRECTION string

	CTE_SCAN_PROP string
	CTE_NAME_PROP string

	ARRAY_INDEX_KEY string
}

func (p PlanProps) ToJSON() []byte {
	marshal, err := json.Marshal(p)
	if err != nil {
		return nil
	}

	return marshal
}

var PropsExported = PlanProps{
	NODE_TYPE_PROP:           NODE_TYPE_PROP,
	EXECUTION_TIME_PROP:      EXECUTION_TIME_PROP,
	ACTUAL_ROWS_PROP:         ACTUAL_ROWS_PROP,
	PLAN_ROWS_PROP:           PLAN_ROWS_PROP,
	ACTUAL_TOTAL_TIME_PROP:   ACTUAL_TOTAL_TIME_PROP,
	ACTUAL_LOOPS_PROP:        ACTUAL_LOOPS_PROP,
	TOTAL_COST_PROP:          TOTAL_COST_PROP,
	PLANS_PROP:               PLANS_PROP,
	STARTUP_COST:             STARTUP_COST,
	PLAN_WIDTH:               PLAN_WIDTH,
	ACTUAL_STARTUP_TIME_PROP: ACTUAL_STARTUP_TIME_PROP,
	RELATION_NAME_PROP:       RELATION_NAME_PROP,
	SCHEMA_PROP:              SCHEMA_PROP,
	ALIAS_PROP:               ALIAS_PROP,
	GROUP_KEY_PROP:           GROUP_KEY_PROP,
	SORT_KEY_PROP:            SORT_KEY_PROP,
	JOIN_TYPE_PROP:           JOIN_TYPE_PROP,
	INDEX_NAME_PROP:          INDEX_NAME_PROP,
	HASH_CONDITION_PROP:      HASH_CONDITION_PROP,

	// computed
	COMPUTED_TAGS_PROP: COMPUTED_TAGS_PROP,

	COSTLIEST_NODE_PROP: COSTLIEST_NODE_PROP,
	LARGEST_NODE_PROP:   LARGEST_NODE_PROP,
	SLOWEST_NODE_PROP:   SLOWEST_NODE_PROP,

	MAXIMUM_COSTS_PROP:         MAXIMUM_COSTS_PROP,
	MAXIMUM_ROWS_PROP:          MAXIMUM_ROWS_PROP,
	MAXIMUM_DURATION_PROP:      MAXIMUM_DURATION_PROP,
	ACTUAL_DURATION_PROP:       ACTUAL_DURATION_PROP,
	ACTUAL_COST_PROP:           ACTUAL_COST_PROP,
	PLANNER_ESTIMATE_FACTOR:    PLANNER_ESTIMATE_FACTOR,
	PLANNER_ESTIMATE_DIRECTION: PLANNER_ESTIMATE_DIRECTION,

	CTE_SCAN_PROP: CTE_SCAN_PROP,
	CTE_NAME_PROP: CTE_NAME_PROP,

	ARRAY_INDEX_KEY: ARRAY_INDEX_KEY,
}

type PlanEnricher struct {
	maxRows     float64
	maxCost     float64
	maxDuration float64
}

func NewPlanEnricher() *PlanEnricher {
	return &PlanEnricher{
		maxRows:     0,
		maxCost:     0,
		maxDuration: 0,
	}
}

func (ps *PlanEnricher) AnalyzePlan(plan map[string]interface{}) {
	ps.processNode(plan)
	plan[MAXIMUM_ROWS_PROP] = ps.maxRows
	plan[MAXIMUM_COSTS_PROP] = ps.maxCost
	plan[MAXIMUM_DURATION_PROP] = ps.maxDuration

	ps.findOutlierNodes(plan)
}

func (ps *PlanEnricher) processNode(node map[string]interface{}) {
	ps.calculatePlannerEstimate(node)
	ps.calculateActuals(node)
	for key, value := range node {
		ps.calculateMaximums(node, key, value)

		if key == PLANS_PROP {
			for _, value := range value.([]interface{}) {
				ps.processNode(value.(map[string]interface{}))
			}
		}
	}
}

func (ps *PlanEnricher) calculateMaximums(node map[string]interface{}, key string, value interface{}) {
	var valueFloat float64
	switch value.(type) {
	case float64:
		valueFloat = value.(float64)
	default:
		return
	}

	if key == ACTUAL_ROWS_PROP && ps.maxRows < valueFloat {
		ps.maxRows = valueFloat
	}

	if key == ACTUAL_COST_PROP && ps.maxCost < valueFloat {
		ps.maxCost = valueFloat
	}

	if key == ACTUAL_DURATION_PROP && ps.maxDuration < valueFloat {
		ps.maxDuration = valueFloat
	}
}

func (ps *PlanEnricher) findOutlierNodes(node map[string]interface{}) {
	node[SLOWEST_NODE_PROP] = false
	node[LARGEST_NODE_PROP] = false
	node[COSTLIEST_NODE_PROP] = false

	if node[ACTUAL_COST_PROP] == ps.maxCost {
		node[COSTLIEST_NODE_PROP] = true
	}
	if node[ACTUAL_ROWS_PROP] == ps.maxRows {
		node[LARGEST_NODE_PROP] = true
	}
	if node[ACTUAL_DURATION_PROP] == ps.maxDuration {
		node[SLOWEST_NODE_PROP] = true
	}

	for key, value := range node {
		if key == PLANS_PROP {
			for _, subNode := range value.([]interface{}) {
				ps.findOutlierNodes(subNode.(map[string]interface{}))
			}
		}
	}
}

func (ps *PlanEnricher) calculatePlannerEstimate(node map[string]interface{}) {
	node[PLANNER_ESTIMATE_FACTOR] = node[ACTUAL_ROWS_PROP].(float64) / node[PLAN_ROWS_PROP].(float64)
	node[PLANNER_ESTIMATE_DIRECTION] = node[PLANNER_ESTIMATE_FACTOR].(float64) >= 1

	if node[PLANNER_ESTIMATE_FACTOR].(float64) < 1 {
		node[PLANNER_ESTIMATE_DIRECTION] = EstimateDirectionOver
		node[PLANNER_ESTIMATE_FACTOR] = node[PLAN_ROWS_PROP].(float64) / node[ACTUAL_ROWS_PROP].(float64)
	}

	if math.IsInf(node[PLANNER_ESTIMATE_FACTOR].(float64), 0) {
		node[PLANNER_ESTIMATE_FACTOR] = 0
	}
	if math.IsNaN(node[PLANNER_ESTIMATE_FACTOR].(float64)) {
		node[PLANNER_ESTIMATE_FACTOR] = 0
	}
}

func (ps *PlanEnricher) calculateActuals(node map[string]interface{}) {
	node[ACTUAL_DURATION_PROP] = node[ACTUAL_TOTAL_TIME_PROP]
	node[ACTUAL_COST_PROP] = node[TOTAL_COST_PROP]

	if node["Plans"] == nil {
		return
	}
	plans := node["Plans"].([]interface{})

	for _, subPlan := range plans {
		sp := subPlan.(map[string]interface{})
		if sp[NODE_TYPE_PROP].(string) != CTE_SCAN_PROP {
			node[ACTUAL_DURATION_PROP] = node[ACTUAL_DURATION_PROP].(float64) - sp[ACTUAL_TOTAL_TIME_PROP].(float64)
			node[ACTUAL_COST_PROP] = node[ACTUAL_COST_PROP].(float64) - sp[TOTAL_COST_PROP].(float64)
		}
	}

	if node[ACTUAL_COST_PROP].(float64) < 0 {
		node[ACTUAL_COST_PROP] = 0
	}

	node[ACTUAL_DURATION_PROP] = node[ACTUAL_DURATION_PROP].(float64) * node[ACTUAL_LOOPS_PROP].(float64)
}
