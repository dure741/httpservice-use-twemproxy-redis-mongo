package thirdurl

import (
	"redis/iriscore/middleware/tracinglog"
	"time"

	iris "github.com/kataras/iris"
	"gitlab.10101111.com/oped/DBMS_LIBS/gorequest"
	log "gitlab.10101111.com/oped/DBMS_LIBS/logrus"
)

func Demo(ctx iris.Context) (string, error) {
	client := gorequest.New().Get("www.cmcm.com").
		//WithContext(ctx.Request().Context()).
		Timeout(5 * time.Second) //.
	//Type("json").
	//Send()

	tracinglog.PutSpanToHeader(ctx, client.Header)
	//resp, body, ierrors := client.End()
	_, body, ierrors := client.End()

	if len(ierrors) != 0 {
		for _, err := range ierrors {
			log.Errorf("gorequest %s err %v", "www.cmcm.com", err)
		}
		return "", ierrors[0]
	}
	return body, nil
}
