package shared

import (
	"bytes"
	"fmt"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"postgres-explain/proto"
	"strings"
	"text/template"
)

type Validate struct {
	PeriodStartFrom *timestamp.Timestamp
	PeriodStartTo   *timestamp.Timestamp
	ClusterName     string
}

func ValidateCommonRequestProps(in Validate) error {
	if in.PeriodStartFrom == nil || in.PeriodStartTo == nil {
		return fmt.Errorf("from-date: %s or to-date: %s cannot be empty", in.PeriodStartFrom, in.PeriodStartTo)
	}

	periodStartFromSec := in.PeriodStartFrom.Seconds
	periodStartToSec := in.PeriodStartTo.Seconds
	if periodStartFromSec > periodStartToSec {
		return fmt.Errorf("from-date %s cannot be bigger then to-date %s", in.PeriodStartFrom, in.PeriodStartTo)
	}

	if in.ClusterName == "" {
		return fmt.Errorf("cluster_name is missing")
	}

	return nil
}

type M map[string]interface{}

func MakeMetrics(mm, t M, durationSec int64) map[string]*proto.MetricValues {
	m := make(map[string]*proto.MetricValues)
	sumNumQueries := interfaceToFloat32(mm["num_queries"])
	m["num_queries"] = &proto.MetricValues{
		Sum:  sumNumQueries,
		Rate: sumNumQueries / float32(durationSec),
	}

	sumNumQueriesWithErrors := interfaceToFloat32(mm["num_queries_with_errors"])
	m["num_queries_with_errors"] = &proto.MetricValues{
		Sum:  sumNumQueriesWithErrors,
		Rate: sumNumQueriesWithErrors / float32(durationSec),
	}

	sumNumQueriesWithWarnings := interfaceToFloat32(mm["num_queries_with_warnings"])
	m["num_queries_with_warnings"] = &proto.MetricValues{
		Sum:  sumNumQueriesWithWarnings,
		Rate: sumNumQueriesWithWarnings / float32(durationSec),
	}

	for k := range commonColumnNames {
		cnt := interfaceToFloat32(mm["m_"+k+"_cnt"])
		sum := interfaceToFloat32(mm["m_"+k+"_sum"])
		totalSum := interfaceToFloat32(mm["m_"+k+"sum"])
		mv := proto.MetricValues{
			Cnt: cnt,
			Sum: sum,
			Min: interfaceToFloat32(mm["m_"+k+"_min"]),
			Max: interfaceToFloat32(mm["m_"+k+"_max"]),
			P99: interfaceToFloat32(mm["m_"+k+"_p99"]),
		}
		if sumNumQueries > 0 && sum > 0 {
			mv.Avg = sum / sumNumQueries
		}
		if sum > 0 && totalSum > 0 {
			mv.PercentOfTotal = sum / totalSum
		}
		if sum > 0 && durationSec > 0 {
			mv.Rate = sum / float32(durationSec)
		}
		m[k] = &mv
	}

	for k := range sumColumnNames {
		cnt := interfaceToFloat32(mm["m_"+k+"_cnt"])
		sum := interfaceToFloat32(mm["m_"+k+"_sum"])
		totalSum := interfaceToFloat32(t["m_"+k+"sum"])
		mv := proto.MetricValues{
			Cnt: cnt,
			Sum: sum,
		}
		if sumNumQueries > 0 && sum > 0 {
			mv.Avg = sum / sumNumQueries
		}
		if sum > 0 && totalSum > 0 {
			mv.PercentOfTotal = sum / totalSum
		}
		if sum > 0 && durationSec > 0 {
			mv.Rate = sum / float32(durationSec)
		}
		m[k] = &mv
	}
	return m
}

func interfaceToFloat32(unk interface{}) float32 {
	switch i := unk.(type) {
	case float64:
		return float32(i)
	case float32:
		return i
	case int64:
		return float32(i)
	default:
		return float32(0)
	}
}

func ProcessQueryWithTemplate(tmplArgs interface{}, arg map[string]interface{}, queryTmpl string) (string, []interface{}, error) {
	var queryBuffer bytes.Buffer
	if tmpl, err := template.New("queryTmpl").Funcs(FuncMap).Parse(queryTmpl); err != nil {
		return "", nil, fmt.Errorf("could not create template: %v", err)
	} else if err = tmpl.Execute(&queryBuffer, tmplArgs); err != nil {
		return "", nil, fmt.Errorf("could not execute template: %v", err)
	}

	return processQuery(queryBuffer, arg)
}

func processQuery(queryBuffer bytes.Buffer, arg map[string]interface{}) (string, []interface{}, error) {
	query, vals, err := sqlx.Named(queryBuffer.String(), arg)
	if err != nil {
		return "", nil, errors.Wrap(err, cannotPrepare)
	}
	query, vals, err = sqlx.In(query, vals...)
	if err != nil {
		return "", nil, errors.Wrap(err, cannotPopulate)
	}
	return query, vals, nil
}

var FuncMap = template.FuncMap{
	"inc":         func(i int) int { return i + 1 },
	"StringsJoin": strings.Join,
}

// workaround to issues in closed PR https://github.com/jmoiron/sqlx/pull/579
func EscapeColons(in string) string {
	return strings.ReplaceAll(in, ":", "::")
}

func EscapeColonsInMap(m map[string][]string) map[string][]string {
	escapedMap := make(map[string][]string, len(m))
	for k, v := range m {
		key := EscapeColons(k)
		escapedMap[key] = make([]string, len(v))
		for i, value := range v {
			escapedMap[key][i] = EscapeColons(value)
		}
	}
	return escapedMap
}
