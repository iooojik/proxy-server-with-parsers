package provider

import (
	"fmt"
	"github.com/iooojik-dev/proxy/mysql"
	"github.com/iooojik-dev/proxy/utility"
	"github.com/pkg/errors"
)

type Proxy struct {
	mysqlClient  *mysql.Client[ProxyRow]
	databaseName string
}

type ProxyRow struct {
	Id           int
	Source       string
	Available    string
	DateCreated  string
	DateModified *string
	Host         string
	Type         string
}

// NewProxyProvider создание провайдера между клиентом mysql и парсерами
func NewProxyProvider() (*Proxy, error) {
	cl, err := mysql.NewClient[ProxyRow]()
	if err != nil {
		return nil, err
	}
	return &Proxy{
		mysqlClient:  cl,
		databaseName: utility.DefaultConfig.Mysql.Database,
	}, nil
}

func (p *Proxy) Insert(items []any) error {
	return errors.New("not implemented")
}

// GetAll получение всех проксей из базы
func (p *Proxy) GetAll() (any, error) {
	return p.mysqlClient.ParseSelectResponse(fmt.Sprintf(`SELECT * FROM %s.proxies;`, p.databaseName))
}
