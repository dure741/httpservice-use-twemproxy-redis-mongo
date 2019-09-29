package iocgo

import (
	"redis/iriscore/config"

	log "gitlab.10101111.com/oped/DBMS_LIBS/logrus"
)

type Initializer interface {
	Init(cfg *config.ApiConf) error
	Close() error
}

var servie_pool []*ServiceItem

type ServiceItem struct {
	initptr Initializer
	name    string
}

func Register(iname string, iner Initializer) {
	item := ServiceItem{initptr: iner, name: iname}
	servie_pool = append(servie_pool, &item)
}

func LaunchEngine(cfg *config.ApiConf) (err error) {
	for _, initfunc := range servie_pool {
		err = initfunc.initptr.Init(cfg)
		if err != nil {
			log.Errorf("init resource[%s] err:%v", initfunc.name, err)
			return err
		}
		log.Infof("init resource[%s] ok", initfunc.name)
	}

	return nil
}

func StopEngine() (err error) {
	for _, initfunc := range servie_pool {
		err = initfunc.initptr.Close()
		if err != nil {
			log.Errorf("close resource[%s] err:%v", initfunc.name, err)
			//return err
		}
		log.Infof("close resource[%s] ok", initfunc.name)
	}

	return nil
}
