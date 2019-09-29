package service

import (
	"sync"

	"redis/iriscore/resource"
)

var (
	once sync.Once
	ser  *DemoService
)

type DemoService struct {
}

func GetSingleTon() *DemoService {
	once.Do(
		func() {
			ser = &DemoService{}
		})

	return ser
}

func (d *DemoService) Demook() string {
	return resource.SigleTon().YourFunction("hello gu long")
	//return "hello iris"
}

func (d *DemoService) Demook2() string {
	return resource.SigleTon().YourFunction("hello jin yong")
	//return "hello iris"
}



func (d *DemoService) GetUserInfo(username string) (string, error) {
	str, err := resource.SigleTon().Hget("username",username,"test","users")
	return str, err
}

func (d *DemoService) UpdateInfo(username, key, value string) (string, error) {
	str, err := resource.SigleTon().Update("username",username, key, value,"test","users")
	return str, err

}

func (d *DemoService) GetAccessInfo(ipaddr,path string)(string,error){
	value:=ipaddr+"->"+path
	str,err:=resource.SigleTon().Hget("ipaddrtopath",value,"test","access")
	return str,err
}

func (d *DemoService)UserInsert(username string)(string,error){
	str,err:=resource.SigleTon().UserInsert(username)
	return str,err

}