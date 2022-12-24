package parsers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/iooojik-dev/proxy/mysql"
	"github.com/iooojik-dev/proxy/provider"
	"github.com/iooojik-dev/proxy/utility"
	"github.com/iooojik/go-logger"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type GeoNodeResponse struct {
	Data []GeoNodeProxy `json:"data"`
}

type GeoNodeProxy struct {
	Ip             string      `json:"ip"`
	Country        string      `json:"country"`
	CreatedAt      time.Time   `json:"created_at"`
	Latency        float64     `json:"latency"`
	Port           string      `json:"port"`
	ResponseTime   int         `json:"responseTime"`
	Speed          int         `json:"speed"`
	UpdatedAt      time.Time   `json:"updated_at"`
	WorkingPercent interface{} `json:"workingPercent"`
}

type GeoNodeParser struct {
	url          string
	databaseName string
	mysqlClient  *mysql.Client[provider.ProxyRow]
}

func NewGeoNodeParser() *GeoNodeParser {
	cl, err := mysql.NewClient[provider.ProxyRow]()
	if err != nil {
		log.Fatal(err)
	}
	return &GeoNodeParser{
		url:          utility.DefaultConfig.Parsers.GeoNode.Url,
		mysqlClient:  cl,
		databaseName: utility.DefaultConfig.Mysql.Database,
	}
}

func (g *GeoNodeParser) Run() {
	// получение проксей из источника
	items, err := g.parse()
	if err != nil {
		log.Fatal(err)
	}
	// сохранение в базу
	for _, p := range items.Data {
		err = g.Insert(p)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (g *GeoNodeParser) parse() (*GeoNodeResponse, error) {
	res, err := g.getResponse()
	if err != nil {
		return nil, err
	}
	return g.parseResponse(res)
}

func (g *GeoNodeParser) getResponse() (*http.Response, error) {
	res, err := http.Get(g.url)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (g *GeoNodeParser) Insert(item any) error {
	prx := item.(GeoNodeProxy)
	table := fmt.Sprintf(`%s.proxies`, g.databaseName)
	query := fmt.Sprintf(`INSERT INTO %s VALUES (NULL, %s, %s, NOW(), NOW(), %s, %s);`, table, "\"proxylist.geonode.com\"", "1", fmt.Sprintf("\"%s:%s\"", prx.Ip, prx.Port), "\"socks5\"")
	insertResult, err := g.mysqlClient.Connection.ExecContext(context.Background(), query)
	if err != nil {
		return err
	}
	if id, e := insertResult.LastInsertId(); e != nil {
		return e
	} else {
		logger.LogPositive(id)
	}
	return nil
}

func (g *GeoNodeParser) GetAll() (any, error) {
	table := fmt.Sprintf(`%s.proxies`, g.databaseName)
	src := fmt.Sprintf(`%s.proxies.src`, g.databaseName)
	return g.mysqlClient.ParseSelectResponse(fmt.Sprintf(`SELECT * FROM %s WHERE %s="proxylist.geonode.com";`, table, src))
}

func (g *GeoNodeParser) parseResponse(response *http.Response) (*GeoNodeResponse, error) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.LogError(err)
		}
	}(response.Body)
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	result := new(GeoNodeResponse)
	if err = json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result, nil
}
