package shared

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"
)

func TestMetricsRepository_Get(t *testing.T) {
	t.Run("query tmpl", func(t *testing.T) {
		tmplArgs := struct {
			PeriodStartFrom int64
			PeriodStartTo   int64
			PeriodDuration  int64
			Dimensions      map[string][]string
			Labels          map[string][]string
			DimensionVal    string
			Group           string
			Totals          bool
		}{
			PeriodStartFrom: 100000000,
			PeriodStartTo:   100000000,
			PeriodDuration:  100,
			DimensionVal:    "123",
			Group:           "queryid",
			Totals:          true,
		}
		var queryBuffer bytes.Buffer
		if tmpl, err := template.New("queryMetricsTmpl").Funcs(FuncMap).Parse(queryMetricsTmpl); err != nil {
		} else if err = tmpl.Execute(&queryBuffer, tmplArgs); err != nil {
		}

		fmt.Println(queryBuffer.String())
	})
}
