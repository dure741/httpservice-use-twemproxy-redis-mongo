package resource

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"
	"encoding/json"
	"github.com/go-redis/redis"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"redis/iriscore/config"
	"redis/iriscore/iocgo"
)

type Users struct {
	Username        string `json:"username" bson:"username"`
	Createdtime     string `json:"createdtime" bson"createdtime"`
	Createdtimeunix int64  `json:"createdtimeunix" bson"createdtimeunix"`
	Deleted         bool   `json:"deleted" bson"deleted"`
}

type Access struct{
	Ipaddr2path string `json:"ipaddrtopath" bson:"ipaddrtopath"`
	Latest string `json:"latest" bson:"latest"`
	Count int64 `json:"count" bson:"count"`
}

var demosrc *Demoresorce
var once sync.Once
var defaultttl int
func SigleTon() *Demoresorce {
	once.Do(func() { demosrc = &Demoresorce{} })
	return demosrc
}

///////////// resource skeleton///////////////

type Demoresorce struct {
	democlient  *Xclient
	Redisclient *redis.Client
	Mongoclient *mgo.Session
}

func (d *Demoresorce) Init(cfg *config.ApiConf) error {
	d.democlient = &Xclient{Addr: cfg.Mysql.DataSource}
	d.Redisclient = redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: "",
		DB:       0,
	})
	d.Mongoclient, _ = mgo.Dial(cfg.MongoDB.Addr)
	defaultttl=cfg.Redis.TTL
	return d.democlient.Connect()
}

func (d *Demoresorce) Close() error {
	d.Redisclient.Close()
	d.Mongoclient.Close()
	return d.democlient.Stop()
}

func (d *Demoresorce) YourFunction(args interface{}) string {
	//skeleton
	return fmt.Sprintf("%v", args)
}

func init() {
	iocgo.Register("Demoresorce pool", SigleTon())
}

//////////// connection resource demo , ignore //////////
type Xclient struct {
	//some resorce
	Addr string
}

func (x *Xclient) Connect() error {
	//some client connection
	// todo:
	return nil
}

func (x *Xclient) Stop() error {
	// close service resource
	return nil
}

//操作redis
//查找hash
func (d *Demoresorce) Hget(key string,value string,db string,c string) (string, error) {
	strmap, err := d.Redisclient.HGetAll(value).Result()
	var copyerr error
	if len(strmap) == 0 {
		if key=="username"{
	    	copyerr = Datacopy(key,value, d.Redisclient,new(Users), d.Mongoclient,db,c)
		}else if key=="ipaddrtopath"{
			copyerr = Datacopy(key,value, d.Redisclient,new(Access), d.Mongoclient,db,c)
		}
		if copyerr != nil {
			return fmt.Sprintf("%s", strmap), copyerr
		} else {
			strmap = d.Redisclient.HGetAll(value).Val()
		}
	}
	return fmt.Sprintf("%s", strmap), err
}

//查找键
func (d *Demoresorce) ReadRedisString(key string) (string, error) {
	str := d.Redisclient.Get(key).Val()
	if str == "" {
		return "", errors.New("Key Not Found")
	}
	return str, nil
}


//删除键
func (d *Demoresorce) DeleteRedisString(key string) (string, error) {
	result, err := d.Redisclient.Del(key).Result()
	return strconv.FormatInt(result, 10), err
}

//从缓存中读取数据的策略
//首先实现几个复用的函数：
//	从mongodb中复制到redis
//	func Datacopy()
func Datacopy(key string,value string, r *redis.Client,data interface{}, m *mgo.Session,db string,c string) error {
	//声明map
	mmap := make(map[string]interface{})
	//从mongo中读取键值

	founderr:=m.DB(db).C(c).Find(bson.M{key: value}).One(data)
	if founderr!=nil{
		return founderr
	}
	fmt.Println(data)
	j, _ := json.Marshal(data)
	json.Unmarshal(j, &mmap)
	fmt.Println(mmap)
	//将mmap写入redis
	err := r.HMSet(fmt.Sprintf("%s",value), mmap).Err()
	// if err != nil {
	// 	fmt.Println("Write redis from map failed!")
	// } else {
	// 	fmt.Println(resp)
	// }

	return err
}

//从接口写入的操作与逻辑：
func (d *Demoresorce) Update(skey, svalue,key string, value interface{},db,c string,) (string, error) {
	d.Redisclient.Del(svalue).Err()
	_,MongoUpdateErr:=d.Mongoclient.DB(db).C(c).Upsert(bson.M{skey: svalue},bson.M{"$set":bson.M{key: value}})
	d.Redisclient.Del(svalue).Err()
	if MongoUpdateErr != nil {
		return "MongoDB Upsert Error!",MongoUpdateErr
	}
	return "MongoDB Upsert Ok!",MongoUpdateErr

}


//Access 函数
func (d *Demoresorce)Access(ipaddr,path string){
	keyname:=ipaddr+"->"+path
	d.Redisclient.Del(keyname).Err()
	Latesttime:=time.Now().Format("2006-01-02 15:04:05")
	d.Mongoclient.DB("test").C("access").Upsert(bson.M{"ipaddrtopath":keyname},bson.M{"$set":bson.M{"latest":Latesttime},"$inc":bson.M{"count":+1}})
	d.Redisclient.Del(keyname).Err()
}


//创建初始化一个新用户的函数
func (d *Demoresorce)UserInsert(username string)(string,error){
	resultcound,_:=d.Mongoclient.DB("test").C("users").Find(bson.M{"username":username}).Count()
	if resultcound!=0{
		return "User Already Exist",errors.New("User Already Exist")
	}
	newuser:=&Users{
		Username: username,
		Createdtime: time.Now().Format("2006-01-02 15:04:05"),
		Createdtimeunix: time.Now().Unix(),
		Deleted: false,
	}
	inserterr:=d.Mongoclient.DB("test").C("users").Insert(newuser)
	if inserterr!=nil{
		return "Insert User Error",inserterr
	}
	return "Insert User Ok",inserterr
}