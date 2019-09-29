package middleware

import (
	"gitlab.10101111.com/oped/DBMS_LIBS/logrus"
	"redis/iriscore/resource"

	iris "github.com/kataras/iris"
)

func DemoMiddleware(ctx iris.Context) {
	logrus.Infof("in demo middleware")
	ctx.Next()
}
func DemoPartyMiddleware(ctx iris.Context) {
	logrus.Infof("in party middleware")
	ctx.Next()
}

func AccessMiddleware(ctx iris.Context){
	ipaddr:=ctx.RemoteAddr()
	path:=ctx.Path()
	resource.SigleTon().Access(ipaddr,path)
	ctx.Next()
}
