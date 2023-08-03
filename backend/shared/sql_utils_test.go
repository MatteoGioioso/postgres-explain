package shared

import (
	"reflect"
	"testing"
)

func TestProcessQueryWithTemplate(t *testing.T) {
	type TmplArgs struct {
		OrderDir string
	}

	type args struct {
		tmplArgs  interface{}
		arg       map[string]interface{}
		queryTmpl string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   []interface{}
		wantErr bool
	}{
		{name: "test", args: args{
			tmplArgs: TmplArgs{OrderDir: "DESC"},
			arg: map[string]interface{}{
				"cluster":   "mycluster",
				"order_by":  "period_start",
				"order_dir": "DESC",
				"limit":     100,
			},
			queryTmpl: `
SELECT id, alias, period_start, query, tracking_id 
FROM plans 
WHERE cluster = :cluster
ORDER BY :order_by {{ .OrderDir }} 
LIMIT :limit
`,
		}, want: "", want1: nil, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := ProcessQueryWithTemplate(tt.args.tmplArgs, tt.args.arg, tt.args.queryTmpl)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessQueryWithTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Log(got, got1)
			if got != tt.want {
				t.Errorf("ProcessQueryWithTemplate() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("ProcessQueryWithTemplate() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestConvertQueryWithParams(t *testing.T) {
	type args struct {
		query  string
		params []interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "query with number",
			args: args{
				query:  "SELECT abalance FROM pgbench_accounts WHERE aid = $1",
				params: []interface{}{3},
			},
			want: "SELECT abalance FROM pgbench_accounts WHERE aid = 3",
		},
		{
			name: "query with string",
			args: args{
				query:  "SELECT abalance FROM pgbench_accounts WHERE aid = $1",
				params: []interface{}{"hello"},
			},
			want: "SELECT abalance FROM pgbench_accounts WHERE aid = 'hello'",
		},
		{
			name: "query with boolean",
			args: args{
				query:  "SELECT abalance FROM pgbench_accounts WHERE aid = $1",
				params: []interface{}{true},
			},
			want: "SELECT abalance FROM pgbench_accounts WHERE aid = true",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertQueryWithParams(tt.args.query, tt.args.params); got != tt.want {
				t.Errorf("ConvertQueryWithParams() = %v, want %v", got, tt.want)
			}
		})
	}
}
