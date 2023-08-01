package shared

import "database/sql"

func ToSqlNullString(val string) sql.NullString {
	if val == "" {
		return sql.NullString{}
	}

	return sql.NullString{
		String: val,
		Valid:  true,
	}
}
