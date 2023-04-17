package config

type App struct {
	Session struct {
		Name      string `yaml:"name"`
		SecretKey string `yaml:"secret_key"`
		MaxAge    int    `yaml:"max_age"`
	} `yaml:"session"`

	AuthMode string `yaml:"auth_mode"`

	//这个结构体使用了yaml标签，表示在将该结构体序列化为yaml格式的文本时，将`db`作为该结构体的标签名
	DB struct {
		Default DB
	} `yaml:"db"`

	LDAP LDAP `yaml:"ldap"`

	Redis struct {
		Default Redis
	} `yaml:"redis"`

	OAuth2 struct {
		AccessTokenExp int            `yaml:"access_token_exp"`
		JWTSignedKey   string         `yaml:"jwt_signed_key"`
		Client         []OAuth2Client `yaml:"client"`
	} `yaml:"oauth2"`
}

type DB struct {
	Type     string `yaml:"type"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

type Redis struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type OAuth2Client struct {
	ID     string  `yaml:"id"`
	Secret string  `yaml:"secret"`
	Name   string  `yaml:"name"`
	Domain string  `yaml:"domain"`
	Scope  []Scope `yaml:"scope"`
}

type Scope struct {
	ID    string `yaml:"id"`
	Title string `yaml:"title"`
}

type LDAP struct {
	URL            string `yaml:"url"`
	SearchDN       string `yaml:"search_dn"`
	SearchPassword string `yaml:"search_password"`
	BaseDN         string `yaml:"base_dn"`
	Filter         string `yaml:"filter"`
}
