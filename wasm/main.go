package main

import (
	"encoding/json"
	"fmt"
	"postgres-explain/core/pkg"
	"runtime/debug"
	"syscall/js"
)

func main() {
	fmt.Println("Starting explain WASM")
	js.Global().Set("explain", explain())
	js.Global().Set("compare", compare())
	<-make(chan bool)
}

func marshalError(expErr pkg.ExplainedError) string {
	marshal, err := json.Marshal(expErr)
	if err != nil {
		panic(err)
	}
	return string(marshal)
}

func explain() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) (ret any) {
		defer func() {
			if r := recover(); r != nil {
				ret = map[string]any{
					"error": marshalError(pkg.ExplainedError{
						Error:   "query explain panic",
						Details: fmt.Errorf("%s", r).Error(),
						Stack:   string(debug.Stack()),
					}),
				}
			}
		}()

		if len(args) != 1 {
			return map[string]any{
				"error": pkg.ExplainedError{Error: "invalid no of arguments passed"},
			}
		}

		plans := args[0].String()
		node, err := pkg.GetRootNodeFromPlans(plans)
		if err != nil {
			return map[string]any{
				"error": marshalError(pkg.ExplainedError{
					Error:   "invalid input: the plan was probably not valid JSON or text",
					Details: err.Error(),
				}),
			}
		}

		pkg.NewPlanEnricher().AnalyzePlan(node)

		statsGather := pkg.NewStatsGather()
		if err := statsGather.GetStatsFromPlans(plans); err != nil {
			return map[string]any{
				"error": marshalError(pkg.ExplainedError{
					Error:   "could not get stats from plan",
					Details: err.Error(),
				}),
			}
		}

		stats := statsGather.ComputeStats(node)
		indexesStats := statsGather.ComputeIndexesStats(node)
		tablesStats := statsGather.ComputeTablesStats(node)
		nodesStats := statsGather.ComputeNodesStats(node)
		jitStats := statsGather.ComputeJITStats()
		triggersStats := statsGather.ComputeTriggersStats()

		summary := pkg.NewSummary().Do(node, stats)

		explained := pkg.Explained{
			Summary:       summary,
			Stats:         stats,
			IndexesStats:  indexesStats,
			TablesStats:   tablesStats,
			NodesStats:    nodesStats,
			JITStats:      jitStats,
			TriggersStats: triggersStats,
		}

		marshalledExplained, err := json.Marshal(explained)
		if err != nil {
			return map[string]any{
				"error": marshalError(pkg.ExplainedError{
					Error:   "could not marshal enriched plan",
					Details: err.Error(),
				}),
			}
		}

		return map[string]any{
			"explained": string(marshalledExplained),
			"error":     nil,
		}
	})
}

func compare() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) (ret any) {
		defer func() {
			if r := recover(); r != nil {
				ret = map[string]any{
					"error": marshalError(pkg.ExplainedError{
						Error:   "plan comparison panic",
						Details: fmt.Errorf("%s", r).Error(),
						Stack:   string(debug.Stack()),
					}),
				}
			}
		}()

		if len(args) != 2 {
			return map[string]any{
				"error": pkg.ExplainedError{Error: "invalid no of arguments passed"},
			}
		}

		planFromArgs := args[0].String()
		planToCompareFromArgs := args[1].String()
		plan := pkg.ExplainedComparison{}
		planToCompare := pkg.ExplainedComparison{}
		if err := json.Unmarshal([]byte(planFromArgs), &plan); err != nil {
			return map[string]any{
				"error": marshalError(pkg.ExplainedError{
					Error:   "could not get plan",
					Details: err.Error(),
				}),
			}
		}
		if err := json.Unmarshal([]byte(planToCompareFromArgs), &planToCompare); err != nil {
			return map[string]any{
				"error": marshalError(pkg.ExplainedError{
					Error:   "could not get plan to compare",
					Details: err.Error(),
				}),
			}
		}

		comparator := pkg.NewComparator(plan, planToCompare)
		comparison, err := comparator.Compare()
		if err != nil {
			return map[string]any{
				"error": marshalError(pkg.ExplainedError{
					Error:   "could compare plans",
					Details: err.Error(),
				}),
			}
		}

		marshalledExplained, err := json.Marshal(comparison)
		if err != nil {
			return map[string]any{
				"error": marshalError(pkg.ExplainedError{
					Error:   "could not marshal comparison",
					Details: err.Error(),
				}),
			}
		}

		return map[string]any{
			"comparison": string(marshalledExplained),
			"error":      nil,
		}
	})
}
