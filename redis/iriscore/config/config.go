package config

import (
	"os"

	"gitlab.10101111.com/oped/DBMS_LIBS/utils"
)

type ApiConf struct {
	Log      LogConf
	Mysql    MysqlConf
	Http     HttpConf
	Uauth    UAuthConf
	Redis    RedisConf   `toml:"redis"`
	MongoDB  MongoDBConf `toml:"mongoDB"`
	Etcd     EtcdConf    `toml:"etcd"`
	ConfPath string
}

type HttpConf struct {
	ServiceID string `toml:"serviceid"`
	HttpAddr  string `toml:"httpaddr"`
	PprofAddr string `toml:"pprofaddr"`
	Timeout   int    `toml:"http_timeout"`
}

type LogConf struct {
	Logpath       string `toml:"logpath,required"`
	AccessLogPath string `toml:"urllogpath,required"`
	TracerLogPath string `toml:"tracerlogpath, required"`
	Loglevel      string `toml:"loglevel"`
	Pidfile       string `toml:"pidfile"`
}

type MysqlConf struct {
	MaxConn    int    `toml:"maxconn"`
	DataSource string `toml:"datasource"`
}

type UAuthConf struct {
	Authurl string `toml:"uauth_url"`
	Project string `toml:"uauth_project"`
}

type RedisConf struct {
	Addr string `toml:"addr"`
	TTL int `toml:"ttl"`
}

type MongoDBConf struct {
	Addr string `toml:"addr"`
}

type EtcdConf struct {
	Addr string `toml:"addr"`
}

func NewConfig(filePath string) (conf *ApiConf, err error) {
	conf = &ApiConf{}
	conf.ConfPath = filePath
	hostname, _ := os.Hostname()
	conf.Http.ServiceID = "SER_" + hostname
	err = utils.LoadTomlCfg(filePath, conf)
	return conf, err
}

func ReloadConf(c *ApiConf, filePath string) error {
	conf := &ApiConf{}
	conf.ConfPath = filePath
	err := utils.LoadTomlCfg(filePath, conf)
	if err != nil {
		return err
	}

	c = conf
	return nil
}
