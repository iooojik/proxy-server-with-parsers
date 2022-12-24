package provider

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestProxy_GetAllProxies(t *testing.T) {
	cl, err := NewProxyProvider()
	defer func(Connection *sql.DB) {
		err := Connection.Close()
		if err != nil {
			assert.NoError(t, err)
		}
	}(cl.mysqlClient.Connection)
	assert.NoError(t, err)
	proxies, err := cl.GetAllProxies()
	assert.NoError(t, err)
	for _, prx := range proxies {
		log.Println(prx)
	}
}
