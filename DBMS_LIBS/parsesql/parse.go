package parsesql

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/pingcap/tidb/ast"
	"github.com/pingcap/tidb/parser"
)

const (
	MAX_LIMIT_COUNT_DEFAULT = 1000000
)

var (
	MAX_LIMIT_COUNT = 10000
)

type Mysqlvisitor struct {
	Tablelist []string
}

func (v *Mysqlvisitor) Enter(in ast.Node) (out ast.Node, skipChildren bool) {
	if haha, ok := in.(*ast.TableName); ok {
		v.Tablelist = append(v.Tablelist, haha.Name.String())
	}
	return in, false
}
func (v *Mysqlvisitor) Leave(in ast.Node) (out ast.Node, ok bool) {
	return in, true
}

func GetTablelist(sqlstmt string) ([]string, error) {
	stmtNode, err := parser.New().ParseOneStmt(sqlstmt, "utf-8", "")
	if err != nil {
		//fmt.Printf("parse error:\n%v\n%s", err, sqlstmt)
		return nil, err
	}
	v := Mysqlvisitor{}
	_, ok := stmtNode.Accept(&v)
	if !ok {
		err = errors.New("get table from parse false")
		//fmt.Printf("gettable from parse false")
		return nil, err
	}
	return v.Tablelist, nil
}

func HasLimit(sqlstmt string) (bool, int, string) {
	stmtNode, err := parser.New().ParseOneStmt(sqlstmt, "utf-8", "")
	if err != nil {
		//fmt.Printf("parse error:\n%v\n%s", err, sqlstmt)
		return false, 0, ""
	}
	selecnode, ok := stmtNode.(*ast.SelectStmt)
	if !ok {
		return false, 0, ""
	}

	if selecnode.Limit == nil {
		return false, 0, ""
	}

	if selecnode.Limit.Count.GetDatum().IsNull() {
		return false, 0, ""
	}

	count := int(selecnode.Limit.Count.GetDatum().GetInt64())
	if count < MAX_LIMIT_COUNT {
		return true, count, sqlstmt
	}
	count = MAX_LIMIT_COUNT
	bakstmt := strings.ToLower(sqlstmt)
	ppos := strings.LastIndex(strings.ToLower(bakstmt), "limit")
	if ppos == -1 {
		return true, count, sqlstmt
	}

	stmtpre := sqlstmt[:ppos]
	if selecnode.Limit.Offset == nil {
		stmtNode.SetText(stmtpre + "limit " + strconv.Itoa(count))
		return true, count, stmtNode.Text()
	} else {
		stmtNode.SetText(stmtpre + "limit " + strconv.Itoa(int(selecnode.Limit.Offset.GetDatum().GetInt64())) + "," + strconv.Itoa(count))
		return true, count, stmtNode.Text()
	}

	return true, count, sqlstmt
}

func SelectFields(sqlstmt string) ([]string, error) {
	stmtNode, err := parser.New().ParseOneStmt(sqlstmt, "utf-8", "")
	if err != nil {
		return nil, err
	}

	selecnode, ok := stmtNode.(*ast.SelectStmt)
	if !ok {
		return nil, nil
	}
	var orignalfield []string
	for _, fd := range selecnode.Fields.Fields {

		if fd.AsName.String() != "" {
			amiin := sqlstmt[fd.Offset:]
			pos := strings.Index(amiin, " as")
			if pos == -1 {
				return nil, fmt.Errorf("sql parse err:%s not has ' as' key", amiin)
			} else {
				orignalf := amiin[:pos]
				orignalf = strings.TrimSpace(orignalf)
				orgs := strings.Split(orignalf, ".")
				if len(orgs) == 1 {
					orignalfield = append(orignalfield, orgs[0])
				} else {
					orignalfield = append(orignalfield, orgs[1])
				}
			}
		} else {
			if fd.Text() == "" {
				return nil, nil
			} else {
				orignalfield = append(orignalfield, fd.Text())
			}
		}
	}
	if len(orignalfield) == 0 {
		return nil, nil
	}
	return orignalfield, nil

}
