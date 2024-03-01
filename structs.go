package main

import "time"

type packetStats struct {
	dropPkts, dupPkts, reordPkts, rcvdPkts int64
	pbdropPkts                             int64
	minRtt, maxRtt, totRtt                 time.Duration
}

type report struct {
	Received int64
	Dropped  int64
	//DroppedPercent        float64
	Duplicates int64
	Reordered  int64
	//ReorderedPercent      float64
	AverageRTT      time.Duration
	LowestRTT       time.Duration
	HighestRTT      time.Duration
	PBQueueDropped  int64
	PBQueueLength   int
	PBQueueCapacity int
}

// provides an incomplete report (missing queue lengths)
func (s packetStats) Report() report {
	var r report
	r.Received = s.rcvdPkts
	r.Dropped = s.dropPkts
	//r.DroppedPercent = float64(s.dropPkts) / float64(s.rcvdPkts+s.dropPkts) * 100
	r.Duplicates = s.dupPkts
	r.Reordered = s.reordPkts
	//r.ReorderedPercent = float64(s.reordPkts) / float64(s.rcvdPkts+s.dropPkts) * 100
	r.LowestRTT = s.minRtt
	r.HighestRTT = s.maxRtt
	r.AverageRTT = time.Duration(float64(s.totRtt) / float64(s.rcvdPkts))
	r.PBQueueDropped = s.pbdropPkts
	return r
}
