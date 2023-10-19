package main

import "time"

type packetStats struct {
	dropPkts, dupPkts, reordPkts, rcvdPkts int64
	pbdropPkts                             int64
	minRtt, maxRtt, totRtt                 time.Duration
}

type report struct {
	Tag                   string
	ReceivedPackets       int64
	DroppedPackets        int64
	DroppedPercent        float64
	DuplicatePackets      int64
	ReorderedPackets      int64
	ReorderedPercent      float64
	AverageRTT            time.Duration
	LowestRTT             time.Duration
	HighestRTT            time.Duration
	PBQueueDroppedPackets int64
	PBQueueLength         int
	PBQueueCapacity       int
}

// provides an incomplete report (missing tag and queue lengths)
func (s packetStats) Report() report {
	var r report
	r.ReceivedPackets = s.rcvdPkts
	r.DroppedPackets = s.dropPkts
	r.DroppedPercent = float64(s.dropPkts) / float64(s.rcvdPkts+s.dropPkts) * 100
	r.DuplicatePackets = s.dupPkts
	r.ReorderedPackets = s.reordPkts
	r.ReorderedPercent = float64(s.reordPkts) / float64(s.rcvdPkts+s.dropPkts) * 100
	r.LowestRTT = s.minRtt
	r.HighestRTT = s.maxRtt
	r.AverageRTT = time.Duration(float64(s.totRtt) / float64(s.rcvdPkts))
	r.PBQueueDroppedPackets = s.pbdropPkts
	return r
}
