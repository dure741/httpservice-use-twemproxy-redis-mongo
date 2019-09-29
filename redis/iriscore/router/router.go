package router

import (
	"redis/iriscore/handler"
	"redis/iriscore/middleware"
	"redis/iriscore/middleware/tracinglog"
)

////////////////////////////////
// router
func (a *API) InitRouter() *API {
	//init middleware here
	a.SetMiddleware(middleware.IAmAlive)
	a.SetMiddleware(handler.RequestLog)
	a.SetMiddleware(middleware.AccessMiddleware)
	a.SetMiddleware(middleware.CheckRatelimit)
	a.SetMiddleware(tracinglog.Tracing)
	//a.SetDone(handler.ResponseLog)
	a.SetDone(tracinglog.FinishSpan)

	////
	users := a.Group("/users", handler.Next)
	users.Get("/", handler.GetUserInfo)
	users.Post("/", handler.UpdateInfo)
	signup:=users.Group("/signup",handler.Next)
	signup.Post("/",handler.Signup)
	access:=a.Group("/access",handler.Next)
	access.Get("/",handler.GetAccessInfo)

	// keepalived api
	a.Get("/do_not_delete.html", nil)

	//global api demo
	{
		a.Get("/demoget", handler.Demo)
		a.Post("/demopost", handler.Demo2)
	}

	// group routing api demo
	{
		p := a.Group("/swordmen_novel", middleware.DemoPartyMiddleware)
		//p.Use(middleware.DemoPartymiddleware)
		p.Get("/gulong", handler.Demo)    //api: http://xxxxx:xxxx/swordmen_novel/gulong
		p.Post("/jinyong", handler.Demo2) //api: http://xxxxx:xxxx/swordmen_novel/jinyong
	}

	return a
}
