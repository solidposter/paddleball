package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/marcboeker/go-duckdb"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type LogMessage struct {
	/*
		{"TimestampUtc":"2025-01-07T11:05:03.507101229Z","level":"INFO","msg":"stats","Tag":"prod1mt103loss1","ReceivedPackets":45,"DroppedPackets":0,"DuplicatePackets":0,"ReorderedPackets":0,"AverageRTT":0.552722,"LowestRTT":0.181409,"HighestRTT":1.292873,"PBQueueDroppedPackets":0,"PBQueueLength":0,"PBQueueCapacity":200}
	*/
	Msg                   string    `json:"msg"`
	Tag                   string    `json:"Tag"`
	TimestampUtc          time.Time `json:"TimestampUtc"`
	ReceivedPackets       int64     `json:"ReceivedPackets"`
	DroppedPackets        int64     `json:"DroppedPackets"`
	DuplicatePackets      int64     `json:"DuplicatePackets"`
	ReorderedPackets      int64     `json:"ReorderedPackets"`
	AverageRTT            float64   `json:"AverageRTT"`
	LowestRTT             float64   `json:"LowestRTT"`
	HighestRTT            float64   `json:"HighestRTT"`
	P90RTT                float64   `json:"P90RTT"`
	P99RTT                float64   `json:"P99RTT"`
	PBQueueDroppedPackets int64     `json:"PBQueueDroppedPackets"`
	PBQueueLength         int       `json:"PBQueueLength"`
	PBQueueCapacity       int       `json:"PBQueueCapacity"`
}

type ReportQuery struct {
	Name              string
	Query             string
	Frequency         time.Duration
	LastExecution     time.Time
	PreparedStatement *sql.Stmt
}

func slogSetup() {
	var logger *slog.Logger
	logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}

func main() {
	duckdbDatabaseFile := flag.String("f", "paddlequack.duckdb", "DuckDB database file name")
	showReport := flag.Bool("r", false, "Show a report every minute")
	flag.Parse()

	// Open DuckDB
	db, err := sql.Open("duckdb", *duckdbDatabaseFile+"?access_mode=read_write")
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error()+"\n")
		panic("Error opening database")
	}
	defer func() {
		// Checkpoint and close the database connection
		//db.Exec("CHECKPOINT")
		db.Close()
	}()
	// Create target table if it does not exist
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS paddleball (
		EventTime timestamp with time zone not null,
		Tag varchar(100) not null,
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
	)`)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error()+"\n")
		panic("Error creating target table")
	}
	insertStmt, err := db.Prepare(`INSERT INTO paddleball (EventTime, Tag, ReceivedPackets, DroppedPackets, DuplicatePackets, ReorderedPackets, AverageRTT, LowestRTT, HighestRTT, P90RTT, P99RTT, PBQueueDroppedPackets, PBQueueLength, PBQueueCapacity)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, nullif(?,0), nullif(?,0), ?, ?, ?)`)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error()+"\n")
		panic("Error preparing insert statement")
	}
	defer insertStmt.Close()
	// Prepare reporting queries
	report := []ReportQuery{}
	if *showReport {
		reportStartTime := time.Now()
		report = []ReportQuery{
			ReportQuery{
				Name: "LastMinute",
				Query: `SELECT SUM(ReceivedPackets)::varchar Received, SUM(DroppedPackets)::varchar Dropped,
					round(AVG(AverageRTT), 4)::varchar AvgRTT, round(MAX(HighestRTT), 4)::varchar MaxRTT, round(MAX(P90RTT), 4)::varchar MaxP90RTT
					FROM paddleball
					WHERE EventTime > (current_timestamp::timestamp - INTERVAL '1 MINUTE')::timestamptz`,
				Frequency:     60 * 1000000000,
				LastExecution: reportStartTime,
			},
			ReportQuery{
				Name: "Last5Minutes",
				Query: `SELECT SUM(ReceivedPackets)::varchar Received, SUM(DroppedPackets)::varchar Dropped,
					round(AVG(AverageRTT), 4)::varchar AvgRTT, round(MAX(HighestRTT), 4)::varchar MaxRTT, round(MAX(P90RTT), 4)::varchar MaxP90RTT
					FROM paddleball
					WHERE EventTime > (current_timestamp::timestamp - INTERVAL '5 MINUTE')::timestamptz`,
				Frequency:     300 * 1000000000,
				LastExecution: reportStartTime,
			},
		}
	}
	for i, _ := range report {
		// Prepare the statements
		q := &report[i]
		q.PreparedStatement, err = db.Prepare(q.Query)
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error()+"\n")
			panic("Error preparing reporting statement")
		}
	}
	defer func() {
		// Close the prepared statements
		for i, _ := range report {
			if report[i].PreparedStatement != nil {
				report[i].PreparedStatement.Close()
			}
		}
	}()
	//
	slogSetup()
	// Register channel for stop processing
	cStop := make(chan os.Signal)
	signal.Notify(cStop, syscall.SIGINT, syscall.SIGTERM)
	// Open stdin and start reading lines from it
	outlines := bufio.NewScanner(os.Stdin)
	for outlines.Scan() {
		var e LogMessage
		textLine := outlines.Text()
		if strings.HasPrefix(textLine, "{") {
			err := json.Unmarshal([]byte(textLine), &e)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error decoding input line, make sure paddleball is logging in JSON format\n")
				break
			}
			if e.Msg == "stats" {
				//fmt.Println(e)
				_, err = insertStmt.Exec(e.TimestampUtc, e.Tag, e.ReceivedPackets, e.DroppedPackets, e.DuplicatePackets, e.ReorderedPackets, e.AverageRTT, e.LowestRTT, e.HighestRTT, e.P90RTT, e.P99RTT, e.PBQueueDroppedPackets, e.PBQueueLength, e.PBQueueCapacity)
				if err != nil {
					fmt.Fprintf(os.Stderr, err.Error()+"\n")
				}
			}
		} else if textLine == "" {
			// Exit in case empty line is read
			break
		}
		// This needs to be in a Goroutine
		reportTime := time.Now()
		for i, _ := range report {
			q := &report[i]
			if q.PreparedStatement != nil && q.LastExecution.Add(q.Frequency).Before(reportTime) {
				q.LastExecution = reportTime
				res, err := q.PreparedStatement.Query()
				if err != nil {
					fmt.Println(err)
					continue
				}
				defer res.Close()
				if res.Next() {
					cols, _ := res.Columns()
					colsNum := len(cols)
					vals := make(map[string]interface{})
					cp := make([]interface{}, colsNum)
					for idx, _ := range cp {
						cp[idx] = new(string)
					}
					err := res.Scan(cp...)
					if err != nil {
						fmt.Println(err)
					}
					for colidx, c := range cols {
						vals[c] = *(cp[colidx].(*string))
					}
					slog.Info(q.Name,
						"values", vals,
					)
				}
			}
		}
		// Below nonblocking check, if the OS signal channel has a termination message
		// The goal here is to exit gracefully, closing DuckDB database gracefully when SIGINT or SIGTERM is sent
		select {
		case _, ok := <-cStop:
			if ok {
				// Break the stdin reading loop
				break
			}
		default:
		}
	}
}
