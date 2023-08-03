package shared

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"strings"
	"text/template"
	"time"
)

const QueryTimeout = 30 * time.Second
const cannotPrepare = "cannot prepare query"
const cannotPopulate = "cannot populate query arguments"
const CannotExecute = "cannot execute query"

func ToSqlNullString(val string) sql.NullString {
	if val == "" {
		return sql.NullString{}
	}

	return sql.NullString{
		String: val,
		Valid:  true,
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

func ConvertQueryWithParams(query string, params []interface{}) (string, error) {
	return SanitizeSQL(query, params...)
}
