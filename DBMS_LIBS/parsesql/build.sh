
go get -d github.com/pingcap/tidb/ast
go get -d github.com/pingcap/tidb/mysql
go get -d github.com/pingcap/tidb/util

go get -d github.com/pingcap/tidb/parser
cd $GOPATH/srcgithub.com/pingcap/tidb/parser
cd goyacc
go build
cd ..
./goyacc/goyacc parser.y

cd -

