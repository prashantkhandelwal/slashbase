package pgqueryengine

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/slashbaseide/slashbase/pkg/queryengines/utils"
)

type pgxConnPoolInstance struct {
	pgxConnPoolInstance *pgxpool.Pool
	LastUsed            time.Time
}

func (pxEngine *PostgresQueryEngine) getConnection(dbConnectionId, host string, port uint16, database, user, password string) (c *pgxpool.Pool, err error) {
	if conn, exists := pxEngine.openConnections[dbConnectionId]; exists {
		pxEngine.mutex.Lock()
		pxEngine.openConnections[dbConnectionId] = pgxConnPoolInstance{
			pgxConnPoolInstance: conn.pgxConnPoolInstance,
			LastUsed:            time.Now(),
		}
		pxEngine.mutex.Unlock()
		return conn.pgxConnPoolInstance, nil
	}
	err = utils.CheckTcpConnection(host, strconv.Itoa(int(port)))
	if err != nil {
		return
	}
	connString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s", host, strconv.Itoa(int(port)), database, user, password)
	pool, err := pgxpool.Connect(context.Background(), connString)
	if err != nil {
		return
	}
	if dbConnectionId != "" {
		pxEngine.mutex.Lock()
		pxEngine.openConnections[dbConnectionId] = pgxConnPoolInstance{
			pgxConnPoolInstance: pool,
			LastUsed:            time.Now(),
		}
		pxEngine.mutex.Unlock()
	}
	return pool, nil
}

func (pxEngine *PostgresQueryEngine) RemoveUnusedConnections() {
	for dbConnID, instance := range pxEngine.openConnections {
		now := time.Now()
		diff := now.Sub(instance.LastUsed)
		if diff.Minutes() > 20 {
			pxEngine.mutex.Lock()
			delete(pxEngine.openConnections, dbConnID)
			pxEngine.mutex.Unlock()
			go instance.pgxConnPoolInstance.Close()
		}
	}
}
