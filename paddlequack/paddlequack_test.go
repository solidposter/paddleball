package main

import (
	"encoding/json"
	"github.com/google/go-cmp/cmp"
	"testing"
	"time"
)

func timeVal(timeString string) time.Time {
	t, _ := time.Parse(time.RFC3339Nano, timeString)
	return t
}

func TestMessageJSONUnmarshal(t *testing.T) {
	/*
		Give some sample messages from paddleball and check if we can Unmarshal them
	*/
	testMessages := []string{
		`{"TimestampUtc":"2025-01-07T11:05:03.507101229Z","level":"INFO","msg":"stats","Tag":"prod1mt103loss1","ReceivedPackets":45,"DroppedPackets":0,"DuplicatePackets":0,"ReorderedPackets":0,"AverageRTT":0.552722,"LowestRTT":0.181409,"HighestRTT":1.292873,"P90RTT":1.192873,"P99RTT":1.205897,"PBQueueDroppedPackets":0,"PBQueueLength":0,"PBQueueCapacity":200}`,
		`{"TimestampUtc":"2025-01-07T11:05:33.159706048Z","level":"INFO","msg":"Starting probe","Tag":"prod1mt103loss1","target":"10.80.5.139:10002"}`,
		`{"TimestampUtc":"2025-01-07T11:05:33.161483383Z","level":"INFO","msg":"Ports active","Tag":"prod1mt103loss1","from":10002,"to":10002}`,
		`{"TimestampUtc":"2025-01-07T11:05:36.160941322Z","level":"INFO","msg":"stats","ReceivedPackets":100,"DroppedPackets":5,"DuplicatePackets":1,"ReorderedPackets":7,"AverageRTT":0.289829,"LowestRTT":0.14008,"HighestRTT":1.205795,"PBQueueDroppedPackets":7,"PBQueueLength":8,"PBQueueCapacity":200,"P90RTT":1.185735,"P99RTT":1.196627}`,
	}
	testResults := []LogMessage{
		LogMessage{
			TimestampUtc:          timeVal("2025-01-07T11:05:03.507101229Z"),
			Msg:                   "stats",
			Tag:                   "prod1mt103loss1",
			ReceivedPackets:       45,
			DroppedPackets:        0,
			DuplicatePackets:      0,
			ReorderedPackets:      0,
			AverageRTT:            0.552722,
			LowestRTT:             0.181409,
			HighestRTT:            1.292873,
			P90RTT:                1.192873,
			P99RTT:                1.205897,
			PBQueueDroppedPackets: 0,
			PBQueueLength:         0,
			PBQueueCapacity:       200,
		},
		LogMessage{
			TimestampUtc: timeVal("2025-01-07T11:05:33.159706048Z"),
			Msg:          "Starting probe",
			Tag:          "prod1mt103loss1",
		},
		LogMessage{
			TimestampUtc: timeVal("2025-01-07T11:05:33.161483383Z"),
			Msg:          "Ports active",
			Tag:          "prod1mt103loss1",
		},
		LogMessage{
			TimestampUtc:          timeVal("2025-01-07T11:05:36.160941322Z"),
			Msg:                   "stats",
			ReceivedPackets:       100,
			DroppedPackets:        5,
			DuplicatePackets:      1,
			ReorderedPackets:      7,
			AverageRTT:            0.289829,
			LowestRTT:             0.14008,
			HighestRTT:            1.205795,
			P90RTT:                1.185735,
			P99RTT:                1.196627,
			PBQueueDroppedPackets: 7,
			PBQueueLength:         8,
			PBQueueCapacity:       200,
		},
	}
	for key, msg := range testMessages {
		var logStruct LogMessage
		err := json.Unmarshal([]byte(msg), &logStruct)
		if err != nil {
			t.Error(err)
		}
		if !cmp.Equal(testResults[key], logStruct) {
			t.Errorf("Wrong object received, got=%s", cmp.Diff(testResults[key], logStruct))
		}
	}
}
