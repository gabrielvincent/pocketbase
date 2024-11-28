//go:build !no_default_driver

package core

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/tursodatabase/go-libsql"
	_ "modernc.org/sqlite"
)

type TursoCredentials struct {
	DBURL   string
	DBToken string
}

func DefaultDBConnect(dbPath string) (*dbx.DB, error) {
	// Note: the busy_timeout pragma must be first because
	// the connection needs to be set to block on busy before WAL mode
	// is set in case it hasn't been already set by another connection.
	pragmas := "?_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)&_pragma=journal_size_limit(200000000)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)&_pragma=temp_store(MEMORY)&_pragma=cache_size(-16000)"

	db, err := dbx.Open("sqlite", dbPath+pragmas)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func TursoDBConnect(dbPath string, creds TursoCredentials) (*dbx.DB, error) {
	// Note: the busy_timeout pragma must be first because
	// the connection needs to be set to block on busy before WAL mode
	// is set in case it hasn't been already set by another connection.
	pragmas := "?_pragma=busy_timeout(100000)&_pragma=journal_mode(WAL)&_pragma=journal_size_limit(200000000)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)&_pragma=temp_store(MEMORY)&_pragma=cache_size(-16000)"

	if strings.HasSuffix(dbPath, "auxiliary.db") {
		return core
	}

	connector, err := libsql.NewEmbeddedReplicaConnector(dbPath, creds.DBURL,
		libsql.WithAuthToken(creds.DBToken),
		libsql.WithReadYourWrites(true),
		libsql.WithSyncInterval(time.Minute),
	)
	if err != nil {
		fmt.Println("Error creating connector:", err)
		os.Exit(1)
	}

	db := dbx.OpenDB(connector, "sqlite", dbPath+pragmas)

	if _, err := connector.Sync(); err != nil {
		return nil, fmt.Errorf("libsql sync error: %v", err)
	}

	return db, nil
}
