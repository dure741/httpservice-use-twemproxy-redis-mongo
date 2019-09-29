package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"redis/iriscore/iocgo"

	"redis/iriscore/config"

	iris "github.com/kataras/iris"

	log "gitlab.10101111.com/oped/DBMS_LIBS/logrus"

	"gitlab.10101111.com/oped/DBMS_LIBS/gorequest"
)

var (
	UAUTH_URL = "" //uauth domain

	UAUTH_URL_AUTHORIZE   = ""
	UAUTH_AUTH_PRJ        = "project=db_search"
	UAUTH_URL_AUTHORIZE_P = "/uauth/query_entitlements_for_3rd_project/"

	HTTP_ERR = errors.New("third part of httpservice resp not 200")
)

type UauthSource struct {
}

func (d *UauthSource) Init(cfg *config.ApiConf) error {
	InitUauthCon(cfg.Uauth.Authurl, cfg.Uauth.Project)
	return nil
}

func (d *UauthSource) Close() error {
	return nil
}

func init() {
	iocgo.Register("uauth service", new(UauthSource))
}

func InitUauthCon(url string, proj string) {
	UAUTH_URL = url
	UAUTH_AUTH_PRJ = fmt.Sprintf("project=%s", proj)
	UAUTH_URL_AUTHORIZE = UAUTH_URL + UAUTH_URL_AUTHORIZE_P
}

func CanIgo(ctx iris.Context, resource, action string) (bool, error) {
	act, err := AuthData(ctx)
	if err != nil {
		return false, err
	}
	return act.IsAuthorized(resource, action), nil
}

func AuthData(ctx iris.Context) (*UauthResp, error) {
	var respbody UauthResp

	client := gorequest.New().Post(UAUTH_URL_AUTHORIZE).
		WithContext(ctx.Request().Context()). // if context has span data, this function can send it to the thirdurl system
		Timeout(2 * time.Second).
		AddCookies(ctx.Request().Cookies()).
		Type("form").
		Send(UAUTH_AUTH_PRJ)

	resp, body, ierrors := client.EndStruct(&respbody)

	if len(ierrors) != 0 {
		mybody := string(body)
		if strings.Contains(mybody, "failure") {
			log.Errorf("AuthData err:%s", mybody)
			return nil, errors.New(mybody)

		}
		for _, err := range ierrors {
			log.Errorf("AuthData gorequest %s err %v", UAUTH_URL_AUTHORIZE, err)
		}
		return nil, ierrors[0]
	}
	if resp.StatusCode != 200 {
		log.Errorf("AuthData gorequest %s not 200:%d", UAUTH_URL_AUTHORIZE, resp.StatusCode)
		return nil, HTTP_ERR
	}

	//log.Debugf("%v", respbody.Data.Myproject)
	for k, v := range respbody.Data.Myproject.ResRoles {
		respbody.Data.Myproject.ResRoles[strings.ToLower(k)] = v
	}
	for k, v := range respbody.Data.Myproject.RsActs {
		respbody.Data.Myproject.RsActs[strings.ToLower(k)] = v
	}

	return &respbody, nil
}

////////////////////////////
// uauth resp struct

type UauthResp struct {
	Msg     string     `json:"msg"`
	Data    DataStruct `json:"data"`
	User    UauthUser  `json:"user"`
	Elapsed string     `json:"elapsed"`
}

type UauthUser struct {
	Userid      string `json:"user_id"`
	Userdomain  string `json:"user_domain"`
	Displayname string `json:"display_name"`
	Useremail   string `json:"user_email"`
}

type DataStruct struct {
	Myproject MyprojectStruct `json:"db_search"` // !!!!!!!!!!!!!!!!! modify this tag to your project name
	Domain    string          `json:"user_domain"`
}

type MyprojectStruct struct {
	RsActs      map[string][]string `json:"role_actions"`
	ResRoles    map[string][]string `json:"resources_roles"`
	RelyOnUtree bool                `json:"rely_on_utree"`
	UserID      string              `json:"user_id"`
}

func (u *UauthResp) IsAuthorized(key string, opertion string) bool {

	r, ok := u.Data.Myproject.ResRoles[key]
	if !ok {
		log.Warnf("thirdurl::authdb::%s:%s no roles", key, opertion)
		return false
	}

	for _, ri := range r {
		a, aok := u.Data.Myproject.RsActs[ri]
		if !aok {
			log.Warnf("thirdurl::authdb::%s:%s:%s action", key, opertion, ri)
			continue
		}

		for _, ai := range a {
			switch {
			case opertion == ai:
				return true

			}
		}
	}
	log.Warnf("thirdurl::authdb::%s:%s action", key, opertion)
	return false
}
