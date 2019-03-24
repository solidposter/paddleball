package main

//
// Copyright (c) 2019 Tony Sarendal <tony@polarcap.org>
//
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
//

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type packetStats struct {
	drops, dups, reords, totPkts int64
	minRtt, maxRtt, totRtt time.Duration
}

type jsonReport struct {
	Source		string
	Sourcetype	string
	Tag		string
	ReceivedPackets		int64
	DroppedPackets		int64
	DuplicatePackets	int64
	ReorderedPackets	int64
	AverageRTT		float64
	LowestRTT		float64
	HighestRTT		float64
}

func statsEngine(rp <-chan payload, global *packetStats,  printJson string) {
	serialNumbers := make(map[int64]int64)	// the expected serial number for each id
	workWindow := []payload{}		// packets to analyze
	feedWindow := []payload{}		// insert packets

	ticker := time.NewTicker(time.Second)
	message := payload{}

	for {
		select {
			case message = <- rp:
				feedWindow = append(feedWindow ,message)
			case <- ticker.C:
				local := process(workWindow, feedWindow, serialNumbers)
				statsUpdate(global,local)

				workWindow = feedWindow		// change feed to work
				feedWindow = []payload{}	// re-init feed

				statsPrint(&local, printJson)
				if printJson == "text" && local.totPkts != 0 {
					fmt.Print(" queue: ",len(rp),"/",cap(rp))
					fmt.Println()
				}
		}
	}
}

func process(workWindow []payload, feedWindow []payload, serialNumbers map[int64]int64) packetStats {
	local := packetStats {}			// local engine info

	for position, message := range workWindow {
		updateRtt(message, &local)

		_, ok := serialNumbers[message.Id]
		if !ok {	// initial packet from this sender ID
			serialNumbers[message.Id] = message.Serial+1
			local.totPkts++
			continue
		}
		if message.Serial == serialNumbers[message.Id] {	// correct order
			local.totPkts++
			local.dups = local.dups + findPacket(serialNumbers, workWindow, feedWindow, position+1, message.Id)	// find duplicates
			serialNumbers[message.Id]++
			continue
		}
		if message.Serial < serialNumbers[message.Id] {		// lower than expected, re-order that already is handled
			local.totPkts++
			continue
		}

		// message.Serial is larger than expected serial.
		// increment til we catch up
		for ; message.Serial > serialNumbers[message.Id]; {	// serial larger, drop or re-order
			matches := findPacket(serialNumbers, workWindow, feedWindow, position, message.Id)
			if matches == 0 {	// packet loss
				local.drops++
				local.totPkts++
				serialNumbers[message.Id]++
				continue
			}
			if matches == 1 {	// re-order
				local.reords++
				local.totPkts++
				serialNumbers[message.Id]++
				continue
			}
			if matches > 1 {	// re-order and duplicates
				local.reords++
				local.dups = local.dups+matches
				local.totPkts++
				serialNumbers[message.Id]++
				continue
			}
		}
		serialNumbers[message.Id]++
	}

	return local
}

func findPacket(serialNumbers map[int64]int64, workWindow []payload, feedWindow []payload, position int, id int64) int64 {
	var n int64	// number of matching packets

	for _,v := range workWindow[position:] {
		if v.Id == id {
			if v.Serial == serialNumbers[v.Id] {
				n++
			}
		}
	}
	for _,v := range feedWindow {
		if v.Id == id {
			if v.Serial == serialNumbers[v.Id] {
				n++
			}
		}
	}
	return n
}

func statsPrint(ei *packetStats, printJson string) {
	if ei.totPkts == 0 {
		return
	}

	if printJson == "text" {
		fmt.Print("packets: ", ei.totPkts)
		fmt.Print(" drops: ", ei.drops)
		fmt.Printf("(%.2f%%) ", float64(ei.drops)/float64(ei.totPkts)*100)
		fmt.Print("re-ordered: ", ei.reords)
		fmt.Printf("(%.2f%%) ", float64(ei.reords)/float64(ei.totPkts)*100)
		fmt.Print("duplicates: ", ei.dups)

		avgRtt := ei.totRtt/time.Duration(ei.totPkts)
		fastest := ei.minRtt-avgRtt	// time below avg rtt
		slowest := ei.maxRtt-avgRtt	// time above avg rtt
		fmt.Print(" avg rtt: ", avgRtt, " fastest: ", fastest, " slowest: +", slowest)
	} else {
		output := jsonReport{}
		output.Source = "PADDLEBALL"
		output.Sourcetype = "PADDLEBALL"
		output.Tag = printJson
		output.DroppedPackets = ei.drops
		output.DuplicatePackets = ei.dups
		output.ReorderedPackets = ei.reords
		output.ReceivedPackets = ei.totPkts

		output.AverageRTT =  float64(ei.totRtt/time.Duration(ei.totPkts)) / 1000000	// avg rtt in ms
		output.LowestRTT = float64(ei.minRtt) / 1000000			// lowest rtt in ms
		output.HighestRTT = float64(ei.maxRtt) / 1000000		// highest rtt in ms

		b, err := json.Marshal(output)
		if err != nil {
			fmt.Println("statsPrint error:",err)
		} else {
			os.Stdout.Write(b)
			fmt.Println()
		}
	}
}

func statsUpdate(global *packetStats, local packetStats) {
	global.drops = global.drops + local.drops
	global.dups = global.dups + local.dups
	global.reords = global.reords + local.reords
	global.totPkts = global.totPkts + local.totPkts
	global.totRtt = global.totRtt + local.totRtt
	if local.minRtt < global.minRtt || global.minRtt == 0 {
		global.minRtt = local.minRtt
	}
	if local.maxRtt > global.maxRtt {
		global.maxRtt = local.maxRtt
	}
}

func updateRtt(message payload, local *packetStats) {
		rtt := message.Rts.Sub(message.Cts)

		local.totRtt = local.totRtt + rtt
		if rtt < local.minRtt || local.minRtt == 0 {
			local.minRtt = rtt
		}
		if rtt > local.maxRtt {
			local.maxRtt = rtt
		}
}
