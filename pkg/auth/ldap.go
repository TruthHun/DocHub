package auth

import (
	"github.com/astaxie/beego/logs"
	"github.com/internet-dev/go-ldap-client"
)

type LdapConfig struct {
	Base    string
	Host    string
	Port    int
	BindDN  string
	BindPwd string
}

func IsAuthorized(userName, password string, conf LdapConfig) (bool, map[string]string, error) {
	logs.Debug("[IsAuthorized] LdapConfig: %#v", conf)
	client := &ldap.LDAPClient{
		Base:         conf.Base,
		Host:         conf.Host,
		Port:         conf.Port,
		UseSSL:       false,
		BindDN:       conf.BindDN,
		BindPassword: conf.BindPwd,
		UserFilter:   "(uid=%s)",
		GroupFilter:  "",
		Attributes:   []string{"givenName", "sn", "mail", "uid"},
	}
	// It is the responsibility of the caller to close the connection
	defer client.Close()

	ok, user, err := client.Authenticate(userName, password)
	if err != nil {
		logs.Error("[IsAuthorized] Error authenticating user %s: %+v", userName, err)
		return false, nil, err
	}
	if !ok {
		logs.Error("Authenticating failed for user %s", userName)
		return false, nil, nil
	}

	return ok, user, nil
}
