package middleware

import (
	"redis/iriscore/handler"
	"redis/thirdurl/auth"

	"github.com/kataras/iris"
)

func CheckSession(ctx iris.Context) {
	data, err := auth.AuthData(ctx)
	if err != nil {
		handler.ResponseErr(ctx, handler.ST_SESSION_OUT, "no session data in uauth")
		return

	}
	ctx.Values().Set(ReqUserKey, data.User.Userid)
	ctx.Values().Set(ReqUserDomain, data.User.Userdomain)
	ctx.Next()
}
