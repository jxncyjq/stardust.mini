package databases

import (
	"bytes"
	"errors"
	"log"
	"text/template"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DBInterface 数据库接口 - 直接使用gorm.DB
type DBInterface *gorm.DB

func NewDBInterface(dbConfig *Config) DBInterface {
	var dbConn DBInterface
	var err error
	if dbConfig.UseMasterSlave {
		dbConn, err = NewMSConn(dbConfig)
		if err != nil {
			panic("Failed to initialize master-slave database connection: " + err.Error())
		}
	} else {
		dbConn, err = NewSingleConn(dbConfig)
		if err != nil {
			panic("Failed to initialize single database connection: " + err.Error())
		}
	}
	return dbConn
}

// NewSingleConn 初始化数据库连接
// mysql fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=%s",user,pwd,host,db,charset)
// postgres fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s sslcert=%s sslkey=%s sslrootcert=%s",
//
//	host, port, user, name, pass, sslMode, SslCert, SslKey, SslRootCert)
func NewSingleConn(c *Config) (DBInterface, error) {
	if nil == c || "" == c.Master {
		return nil, errors.New("config or config.Url can not be null")
	}

	var conn *gorm.DB
	var err error

	switch c.DbType {
	case "mysql":
		conn, err = gorm.Open(mysql.Open(c.Master), &gorm.Config{})
	case "postgres":
		conn, err = gorm.Open(postgres.Open(c.Master), &gorm.Config{})
	default:
		conn, err = gorm.Open(mysql.Open(c.Master), &gorm.Config{})
	}

	if err != nil {
		log.Println("failed to initializing db connection:", err)
		return nil, err
	}

	// 获取底层sql.DB以设置连接池参数
	sqlDB, err := conn.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(c.MaxIdle)
	sqlDB.SetMaxOpenConns(c.MaxConn)

	return conn, nil
}

// NewMSConn 初始化主从数据库连接, master不能为空，slaves可以为空
func NewMSConn(c *Config) (DBInterface, error) {
	if nil == c || "" == c.Master {
		return nil, errors.New("config or config.Url can not be null")
	}

	// GORM 本身不直接支持主从配置，这里只创建主库连接
	// 如果需要主从支持，可以考虑使用GORM的插件或自定义实现
	log.Println("Warning: Master-Slave configuration not fully supported in GORM, using master connection only")

	return NewSingleConn(c)
}

//func GetMySqlDB() DBInterface {
//	return dbConn
//}
//
//func GetDao() BaseDao {
//	return NewBaseDao(dbConn)
//}

// GetConnStr 检查传入字典和类型，返回数据库连接字符串，
// @params c map[string]interface{}{ "host":"127.0.0.1","port":3306,"user":"root","passwd":"123456","dbname":"csv"}
// @param t string "mysql" or "postgres"
// @return string, error
func GetConnStr(c map[string]interface{}, t string) (string, error) {
	var constr bytes.Buffer
	var tpr *template.Template
	var err error
	switch t {
	case "mysql":
		{
			tmpl := `{{.user}}:{{.pwd}}@tcp({{.host}}:{{.port}})/{{.db}}?charset=utf8mb4&parseTime=true&loc=Local`
			tpr, err = template.New("config").Parse(tmpl)
		}
	case "postgres":
		{
			tmpl := `host={{.host}} port={{.port}} user={{.user}} dbname={{.db}} password={{.pwd}} sslmode=disable`
			tpr, err = template.New("config").Parse(tmpl)
		}
	default:
		return "", errors.New("unsupported db type")
	}
	// 错误检查
	if err != nil {
		return "", err
	}
	if err = tpr.Execute(&constr, c); err != nil {
		return "", err
	}
	return constr.String(), nil
}
