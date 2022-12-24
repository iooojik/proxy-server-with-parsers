package mysql

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewClient(t *testing.T) {
	cl, err := NewClient()
	defer func(Connection *sql.DB) {
		e := Connection.Close()
		if e != nil {
			assert.NoError(t, e)
		}
	}(cl.Connection)
	assert.NoError(t, err)
}
