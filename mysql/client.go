package mysql

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iooojik-dev/proxy/utility"
	"github.com/iooojik/go-logger"
	"time"
)

const (
	driver = "mysql"
)

type Clienter interface {
	GetAll() (any, error)
	Insert(item any) error
}

type Client[T any] struct {
	Connection *sql.DB
}

// NewClient создание клиента mysql
func NewClient[T any]() (*Client[T], error) {
	config := utility.DefaultConfig.Mysql
	db, err := sql.Open(driver, fmt.Sprintf("%s:%s@tcp(%s)/%s", config.User, config.Password, config.Host, config.Database))
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &Client[T]{
		Connection: db,
	}, nil
}

func (c *Client[T]) ParseSelectResponse(query string) ([]*T, error) {
	rows, err := c.Connection.Query(query)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		if e := rows.Close(); e != nil {
			logger.LogError(e)
		}
	}(rows)
	colNames, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	cols := make([]string, len(colNames))
	colPtrs := make([]interface{}, len(colNames))
	for i := 0; i < len(colNames); i++ {
		colPtrs[i] = &cols[i]
	}
	items := make([]*T, 0)
	var myMap = make(map[string]string)
	for rows.Next() {
		scanErr := rows.Scan(colPtrs...)
		if scanErr != nil {
			return nil, scanErr
		}
		for i, col := range cols {
			myMap[colNames[i]] = col
		}
		rowItem := new(T)
		data, e := json.Marshal(myMap)
		if e != nil {
			return nil, e
		}
		e = json.Unmarshal(data, rowItem)
		if e != nil {
			return nil, e
		}
		items = append(items, rowItem)
	}
	return items, nil
}
