package model

import (
	"context"
	"errors"
	"oauthsso/config"
	"oauthsso/pkg/ldap"
)

type User struct {
	ID       int    `gorm:"primary_key" json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// 获取用户表名称
func (u *User) TableName() string {
	return "user"
}

func (u *User) Authentication(ctx context.Context, username, password string) (userID string, err error) {
	if config.Get().AuthMode == "ldap" {
		userID, err = ldap.UserAuthentication(username, password)
		return
	}

	//用数据库验证方式
	if config.Get().AuthMode == "db" {
		// write your own user authentication logic
		// like:
		// DB().WithContext(ctx).Where("name = ? AND password = ?", username, password).First(u)
		// userID = u.ID
		if username != "admin" || password != "admin" {
			//test account: admin admin
			err = errors.New("用户密码错误")
			return
		}

		userID = username
		return
	}
	return
}
