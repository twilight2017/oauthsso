package ldap

import (
	"crypto/tls"
	"fmt"
	ldap "github.com/go-ldap/ldap/v3"
	"log"
	"net"
	"net/url"
	"oauthsso/config"
	"strings"
)

type Session struct {
	ldapCfg  config.LDAP
	ldapConn *ldap.Conn
}

func NewSession(ldapCfg config.LDAP) *Session {
	return &Session{
		ldapCfg: ldapCfg,
	}
}

//yaml文件中支持多种ldap地址配置方式，这里进行地址格式化处理
//格式化为protocol://host:port的格式，支持ldap和ldaps协议
func formatURL(ldapURL string) (string, error) {
	var protocol, hostport string
	_, err := url.Parse(ldapURL)
	if err != nil {
		return "", fmt.Errorf("parse Ldap Host ERR: %s", err)
	}
	if strings.Contains(ldapURL, "://") {
		splitLdapURL := strings.Split(ldapURL, "://")
		protocol, hostport = splitLdapURL[0], splitLdapURL[1]
		if !((protocol == "ldap") || (protocol == "ldaps")) {
			return "", fmt.Errorf("unknown ldap protocol")
		}
	} else {
		hostport = ldapURL
		protocol = "ldap"
	}

	if strings.Contains(hostport, ":") {
		_, port, err := net.SplitHostPort(hostport)
		if err != nil {
			return "", fmt.Errorf("illegal ldap url, error: %v", err)
		}
		if port == "636" {
			protocol = "ldaps"
		}
	} else {
		switch protocol {
		case "ldap":
			hostport = hostport + ":389"
		case "ldaps":
			hostport = hostport + ":636"
		}
	}
	fLdapURL := protocol + "://" + hostport
	return fLdapURL, nil
}

//open session
//should invoke Close for each Open call
func (s *Session) Open() error {
	ldapURL, err := formatURL(s.ldapCfg.URL)
	if err != nil {
		return err
	}
	splitLdapURL := strings.Split(ldapURL, "://")

	protocol, hostport := splitLdapURL[0], splitLdapURL[1]
	host, _, err := net.SplitHostPort(hostport)
	if err != nil {
		return err
	}

	log.Println(ldapURL)

	switch protocol {
	case "ldap":
		l, err := ldap.Dial("tcp", hostport)
		if err != nil {
			return err
		}
		s.ldapConn = l
	case "ldaps":
		l, err := ldap.DialTLS("tcp", hostport, &tls.Config{ServerName: host, InsecureSkipVerify: true})
		if err != nil {
			return err
		}
		s.ldapConn = l
	}
	return nil
}

//close current session
func (s *Session) Close() {
	if s.ldapConn != nil {
		s.ldapConn.Close()
	}
}

func UserAuthentication(username, password string) (string, error) {
	s := NewSession(config.Get().LDAP)
	if err := s.Open(); err != nil {
		return "", err
	}
	defer s.Close()

	// First bind with a read only user
	if err := s.ldapConn.Bind(s.ldapCfg.SearchDN, s.ldapCfg.SearchPassword); err != nil {
		return "", err
	}

	//Search for the given username
	searchRequest := ldap.NewSearchRequest(
		s.ldapCfg.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(s.ldapCfg.Filter, ldap.EscapeFilter(username)),
		[]string{"dn"},
		nil,
	)

	sr, err := s.ldapConn.Search(searchRequest)
	if err != nil {
		return "", err
	}

	if len(sr.Entries) != 1 {
		return "", fmt.Errorf("用户不存在或不唯一")
	}

	userdn := sr.Entries[0].DN

	//Bind as the user to verfify their password
	// 这段代码是对LDAP连接进行用户绑定，即使用指定的用户名和密码对LDAP连接进行认证
	//该操作需要在LDAP服务器上已经存在一个对应的账号，并且该账号具有登录权限
	if err := s.ldapConn.Bind(userdn, password); err != nil {
		return "", err
	}

	//Rebind as the read only user for any further queries
	if err := s.ldapConn.Bind(s.ldapCfg.SearchDN, s.ldapCfg.SearchPassword); err != nil {
		return "", err
	}

	return username, nil
}
