package parsesql

import (
	"fmt"
	"testing"

	"github.com/pingcap/tidb/parser"
)

func TestP(t *testing.T) {
	sql := "SELECT  emp_no, first_name, last_name " +
		"FROM employees  inner join cccc on cccc.id = employees.id inner join stusent on cccc.id = student.id  " +
		"where last_name='Aamodt' and gender='F' and birth_date > '1960-01-01' limit 50"

	sqlParser := parser.New()
	stmtNodes, err := sqlParser.Parse(sql, "utf-8", "")
	if err != nil {
		fmt.Printf("parse error:\n%v\n%s", err, sql)
		t.Fatalf("err")
	}
	v := Mysqlvisitor{}
	for _, stmtNode := range stmtNodes {
		stmtNode.Accept(&v)
	}
	fmt.Println("-===================")
	for pos, sql := range v.Tablelist {
		fmt.Printf("tablelist:%d:%s\n", pos, sql)
	}
	if HasLimit(sql) {
		fmt.Printf("parse has limit:%s\n", sql)
	} else {
		fmt.Printf("parse has no limit:%s\n", sql)
	}

	sql = "select * from haha where 1=1"
	if HasLimit(sql) {
		fmt.Printf("parse has limit:%s\n", sql)
	} else {
		fmt.Printf("parse has no limit:%s\n", sql)
	}
	sql = "select * from haha where  id in (select id from hehe limit 50)"
	if HasLimit(sql) {
		fmt.Printf("parse has limit:%s\n", sql)
	} else {
		fmt.Printf("parse has no limit:%s\n", sql)
	}

	sql = "update helll_in_u set name='hehe'"
	ret, err := GetTablelist(sql)
	if err != nil {
		fmt.Printf("parse error:%v\n%s", err, sql)
	}

	fmt.Println("-===================")
	for pos, sqli := range ret {
		fmt.Printf("tablelist->>%d:%s\n", pos, sqli)
	}

	fmt.Println("-===================")
	fmt.Println("-===================")
	fmt.Println("-===================")
	fmt.Println("-===================")
	ret, err = GetTablelist("delete from hello_world where id =3 ")
	if err != nil {
		fmt.Printf("parse error:%v, %s \n", err, "delete from hello_world where id =3")
		t.Fatalf("err")
	}

	fmt.Println("-===================")
	for pos, sqli := range ret {
		fmt.Printf("tablelist->>%d:%s\n", pos, sqli)
	}

}
