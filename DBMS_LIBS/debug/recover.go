package debug

import (
	log "gitlab.10101111.com/oped/DBMS_LIBS/logrus"
)

func ProtectPanic() {
	if err := recover(); err != nil {
		log.Errorf("!!!!!!!! recover from panic %v", err)
	}
}
