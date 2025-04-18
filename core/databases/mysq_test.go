package databases

import (
	"bytes"
	"testing"
	"text/template"
)

func TestEntity_Create(t *testing.T) {
	config := map[string]interface{}{
		"host":   "127.0.0.1",
		"port":   3306,
		"user":   "root",
		"passwd": "123456",
		"dbname": "csv",
	}

	tmpl := `{{.User}}:{{.Passwd}}@tcp({{.Host}}:{{.Port}})/{{.Dbname}}?charset=utf8mb4&parseTime=true&loc=Local`
	r, err := template.New("config").Parse(tmpl)
	if err != nil {
		panic(err)
	}
	var constr bytes.Buffer
	_ = r.Execute(&constr, config)

	c := Config{
		ShowSql: true,
		Master:  constr.String(),
	}
	conn, err := NewConn(&c)
	if err != nil {
		t.Error(err)
	}
	createTable := `create table if not exists test_trans(id int)`
	result, err := conn.Query(createTable)
	if err != nil {
		t.Error(err)
	}
	t.Log(result)

}
