//go:build cgo

/*
	DuckDB built in support
	DuckDB requires CGO, so include this file if build with CGO is requested
*/

package main

import (
	"database/sql"
	"errors"
	_ "github.com/marcboeker/go-duckdb"
	"time"
)

type QuackStats struct {
	DB                                  *sql.DB
	insertLocalStats, insertGlobalStats *sql.Stmt
	Tag                                 string
}

func validateQuackDatabaseFileName(databaseFileName string) bool {
	// TODO: Check somehow that the database file name is actually valid file name
	return true
}

func createQuackStructure(quack *QuackStats) error {
	// Create target table if it does not exist
	columns := `(
		EventTime timestamp with time zone not null,
		Tag varchar(100),
		ReceivedPackets bigint not null,
		DroppedPackets bigint not null,
		DuplicatePackets bigint not null,
		ReorderedPackets bigint not null,
		AverageRTT double not null,
		LowestRTT double not null,
		HighestRTT double not null,
		P90RTT double,
		P99RTT double,
		PBQueueDroppedPackets bigint not null,
		PBQueueLength int not null,
		PBQueueCapacity int not null
	)`
	_, err := quack.DB.Exec("CREATE TABLE IF NOT EXISTS measurement " + columns)
	if err != nil {
		return err
	}
	_, err = quack.DB.Exec("CREATE TABLE IF NOT EXISTS aggregation " + columns)
	if err != nil {
		return err
	}
	return nil
}

func NewQuack(databaseFileName string) (*QuackStats, error) {
	quack := QuackStats{}
	// Open DuckDB
	if !validateQuackDatabaseFileName(databaseFileName) {
		return nil, errors.New("Invalid database file name")
	}
	connectionString := databaseFileName + "?access_mode=read_write"
	var err error
	quack.DB, err = sql.Open("duckdb", connectionString)
	if err != nil {
		return nil, err
	}
	err = createQuackStructure(&quack)
	if err != nil {
		return nil, err
	}
	// Prepare insert statements
	quack.insertLocalStats, err = quack.DB.Prepare(`INSERT INTO measurement (EventTime, Tag, ReceivedPackets, DroppedPackets, DuplicatePackets, ReorderedPackets, AverageRTT, LowestRTT, HighestRTT, P90RTT, P99RTT, PBQueueDroppedPackets, PBQueueLength, PBQueueCapacity)
		VALUES (?, nullif(?, ''), ?, ?, ?, ?, ?, ?, ?, nullif(?,0), nullif(?,0), ?, ?, ?)`)
	if err != nil {
		return nil, err
	}
	quack.insertGlobalStats, err = quack.DB.Prepare(`INSERT INTO aggregation (EventTime, Tag, ReceivedPackets, DroppedPackets, DuplicatePackets, ReorderedPackets, AverageRTT, LowestRTT, HighestRTT, P90RTT, P99RTT, PBQueueDroppedPackets, PBQueueLength, PBQueueCapacity)
		VALUES (?, nullif(?, ''), ?, ?, ?, ?, ?, ?, ?, nullif(?,0), nullif(?,0), ?, ?, ?)`)
	if err != nil {
		return nil, err
	}
	//
	return &quack, nil
}

func (q *QuackStats) Close() {
	// Resource cleanup
	if q.insertLocalStats != nil {
		q.insertLocalStats.Close()
	}
	if q.insertGlobalStats != nil {
		q.insertGlobalStats.Close()
	}
	if q.DB != nil {
		q.DB.Close()
	}
}

func (q *QuackStats) StoreReport(e *report, isMeasurement bool) error {
	var stmt *sql.Stmt
	if isMeasurement {
		stmt = q.insertLocalStats
	} else {
		stmt = q.insertGlobalStats
	}
	AverageRTT := float64(e.AvgRTT) / 1000000
	LowestRTT := float64(e.LowRTT) / 1000000
	HighestRTT := float64(e.HighRTT) / 1000000
	P90RTT := float64(e.P90RTT) / 1000000
	P99RTT := float64(e.P99RTT) / 1000000
	_, err := stmt.Exec(time.Now(), q.Tag, e.Received, e.Drops, e.Dups, e.Reordered, AverageRTT, LowestRTT, HighestRTT, P90RTT, P99RTT, e.PBQueueDrops, e.PBQueueLen, e.PBQueueCap)
	return err
}
