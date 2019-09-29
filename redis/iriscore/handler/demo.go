package handler

import (
	"github.com/kataras/iris"

	"redis/iriscore/service"

	log "gitlab.10101111.com/oped/DBMS_LIBS/logrus"
)

func Demo(ctx iris.Context) {
	retdata := service.GetSingleTon().Demook()
	ReponseOk(ctx, retdata)
	log.Debugf("demo reponse ok")
}

func Demo2(ctx iris.Context) {
	retdata := service.GetSingleTon().Demook2()
	ReponseOk(ctx, retdata)
	log.Debugf("demo reponse ok")
}



func Next(ctx iris.Context) {
	ctx.Next()
}

func GetUserInfo(ctx iris.Context) {
	result, err := service.GetSingleTon().GetUserInfo(ctx.URLParam("username"))
	if err != nil {
		ctx.Text("Get User Information Error")
		log.Errorf("Get User Information Error")
		return
	}
	ctx.Text(result)
	log.Infof("GOt User Information : %s", result)
}

func UpdateInfo(ctx iris.Context) {
	result, err := service.GetSingleTon().UpdateInfo(ctx.URLParam("username"),ctx.URLParam("key"),ctx.URLParam("value"))
	if err != nil {
		ctx.Text("Update User Information Error")
		log.Errorf("Update User Information Error")
		return
	}
	ctx.Text(result)
	log.Infof("Update User Information : %s", result)
}


func GetAccessInfo(ctx iris.Context){
	result,err:=service.GetSingleTon().GetAccessInfo(ctx.URLParam("ipaddr"),ctx.URLParam("path"))
	if err!=nil{
		ctx.Text("Get Access Information Error")
		log.Errorf("Get Access Information Error")
	}
	ctx.Text(result)
	log.Infof("Get Access Information : %s",result)
}

func Signup(ctx iris.Context){
	result,err:=service.GetSingleTon().UserInsert(ctx.URLParam("username"))
	if err!=nil{
		ctx.Text("Insert User Error")
		log.Errorf("Insert User Error")
	}
	ctx.Text(result)
	log.Infof("Insert User Ok")
}
