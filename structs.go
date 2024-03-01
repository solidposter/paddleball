package main

import "time"

type packetStats struct {
	dropPkts, dupPkts, reordPkts, rcvdPkts int64
	pbdropPkts                             int64
	minRtt, maxRtt, totRtt                 time.Duration
}

type report struct {
	Received int64
	Drops    int64
	//DroppedPercent        float64
	Dups      int64
	Reordered int64
	//ReorderedPercent      float64
	AvgRTT       time.Duration
	LowRTT       time.Duration
	HighRTT      time.Duration
	PBQueueDrops int64
	PBQueueLen   int
	PBQueueCap   int
}

// provides an incomplete report (missing queue lengths)
func (s packetStats) Report() report {
	var r report
	r.Received = s.rcvdPkts
	r.Drops = s.dropPkts
	//r.DroppedPercent = float64(s.dropPkts) / float64(s.rcvdPkts+s.dropPkts) * 100
	r.Dups = s.dupPkts
	r.Reordered = s.reordPkts
	//r.ReorderedPercent = float64(s.reordPkts) / float64(s.rcvdPkts+s.dropPkts) * 100
	r.LowRTT = s.minRtt
	r.HighRTT = s.maxRtt
	r.AvgRTT = time.Duration(float64(s.totRtt) / float64(s.rcvdPkts))
	r.PBQueueDrops = s.pbdropPkts
	return r
}
