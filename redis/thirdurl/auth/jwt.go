package auth

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gitlab.10101111.com/oped/DBMS_LIBS/debug"

	iris "github.com/kataras/iris"

	log "gitlab.10101111.com/oped/DBMS_LIBS/logrus"

	"gitlab.10101111.com/oped/DBMS_LIBS/gorequest"
)

type JwtToken struct {
	Token string `json:"token"`
}

func VerifyJwtToken(ctx iris.Context) (string, error) {
	auth_header := ctx.Request().Header.Get("Authorization")
	if auth_header == "" {
		return "", fmt.Errorf("not found jwt token")
	}

	authHeaderParts := strings.Split(auth_header, " ")
	if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != "bearer" {
		log.Warnf("%s token verification failed", auth_header)
		return "", fmt.Errorf("Authorization header format must be Bearer")
	}
	token := authHeaderParts[1]

	//GET
	user := &JwtToken{
		Token: token,
	}
	data, err := json.Marshal(user)
	if err != nil {
		log.Errorf("json.marshal failed, err:", err)
		return "", err
	}
	authurl := UAUTH_URL + "/verify_token/"
	client := gorequest.New().Post(authurl).
		WithContext(ctx.Request().Context()).
		Timeout(5 * time.Second).
		Type("json").
		Send(string(data))

	resp, body, ierrors := client.End()

	if len(ierrors) != 0 {
		for _, err := range ierrors {
			log.Errorf("VerifyToken gorequest %s err %v", authurl, err)
		}
		return "", ierrors[0]
	}
	if resp.StatusCode != 200 {
		log.Errorf("VerifyToken gorequest %s not 200:%d", authurl, resp.StatusCode)
		return "", HTTP_ERR
	}

	var retcode map[string]interface{}
	err = json.Unmarshal([]byte(body), &retcode)
	if err != nil {
		log.Errorf("VerifyToken gorequest %s not json:%s", authurl, string(body))
		return "", err
	}
	return debug.GetValueString("user", retcode), nil
}
