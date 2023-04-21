package model

import (
	"fmt"
	"gorm.io/driver/mysql"
	//"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"oauthsso/config"
	"time"
)

var db *gorm.DB

//该函数用于返回数据库连接，可以在yaml文件里配置连接数据库的类型
func DB() *gorm.DB {
	if db != nil {
		return db
	}

	var err error
	cfg := config.Get().DB.Default
	switch cfg.Type {
	// dsn := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	case "mysql":
		//dsn用字符串格式描述连接类型
		dsn := fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
			cfg.User,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.DBName,
		)
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			//将日志记录器的日志模式设置为静默模式
			Logger: logger.Default.LogMode(logger.Silent),
		})
	case "postgresql":
		// dsn := fmt.Sprintf(
		// 	"host=%s user=%s password=%s dbname = %s port=%d sslmodel=disable TimeZone=Asia/Shanghai",
		// 	cfg.Host,
		// 	cfg.User,
		// 	cfg.Password,
		// 	cfg.DBName,
		// 	cfg.Port,
		// )
		// db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		// 	Logger: logger.Default.LogMode(logger.Silent),
		// })
	}

	if err != nil {
		//将错误信息记录到日志并使程序终止运行的方法
		//该方法会打印错误信息，然后调用'os.Exit(1)'来终止程序的运行
		log.Fatal(err)
	}

	//GORM使用database/sql 维护连接池
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	//连接池可以帮助减少应用程序与数据库建立连接的开销，同时保证有足够的连接可以使用
	sqlDB.SetMaxIdleConns(10)           //设置最大空闲连接数目
	sqlDB.SetMaxOpenConns(100)          //设置连接池中的最大连接数
	sqlDB.SetConnMaxLifetime(time.Hour) //设置单个连接的最大生存时间
	return db
}

//以下代码定义了基础模型的结构体，这个结构体通常会被应用作为应用程序中所有模型的基类，包含了模型所需要的必需字段
type model struct {
	ID        uint       `gorm:"primary_key" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}
