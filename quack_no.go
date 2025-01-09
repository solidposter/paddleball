//go:build !cgo

/*
	Dummy structures for building smaller paddleball without DuckDB
	DuckDB requires CGO, so include this file if build without CGO is requested
*/

package main

type QuackStats struct {
	Tag string
}

func NewQuack(databaseFileName string) (*QuackStats, error) {
	quack := QuackStats{}
	return &quack, nil
}

func (q *QuackStats) Close() {
}

func (q *QuackStats) StoreReport(e *report, isMeasurement bool) error {
	return nil
}
