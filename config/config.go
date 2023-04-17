package config

import (
	"flag"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

var cfg APP

func SetUp() {
	// 这行代码是使用flag包声明一个名为"config"的命令行参数，其默认值为"/etc/oauthsso/config.yaml"
	//"the absolute path of config.yaml"是参数的详细描述信息
	path := flag.String("config", "/etc/oauthsso/config.yaml", "the absolute path of config.yaml")
	flag.Parse()

	content, err := ioutil.ReadFile(*path)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// 将yaml格式的配置文件内容解析到一个cfg结构体变量中
	err = yaml.Unmarshal(content, &cfg)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
