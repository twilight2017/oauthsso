package config

import (
	"strings"
)

func Get() *APP {
	return &cfg
}

//根据clientId查找对应的client
func GetOAuth2Client(clientID string) (cli OAuth2Client) {
	for _, v := range cfg.OAuth2.Client {
		if v.ID == clientID {
			cli = v //注意go语法直接指定了返回变量的名称
		}
	}
	return
}

// 将权限列表按整个字符串返回
func ScopeJoin(scope []Scope) string {
	var s []string
	for _, sc := range scope {
		s = append(s, sc.Title)
	}
	return strings.Join(s, ",")
}

func ScopeFilter(clientID string, scope string) (s []Scope) {
	cli := GetOAuth2Client(clientID)
	//获得权限列表
	s1 := strings.Split(scope, ",")
	// 双重循环判断当前用户允许的权限范围
	for _, str := range s1 {
		for _, sc := range cli.Scope {
			if str == sc.ID {
				s = append(s, sc)
			}
		}

	}
	return //最后返回结果是整个Scope列表
}
