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
	"fmt"
	"time"
)

type engineStats struct {
	numClients, rate int
	drops, dups, reords, totPkts int64
	minRtt, maxRtt, totRtt time.Duration
}

func statsEngine(rp <-chan payload, rate int, numclients int) {
	serialMap := make(map[int64]int64)
	workWindow := []payload{}	// analyze packets
	feedWindow := []payload{}	// insert packets

	ticker := time.NewTicker(time.Second)
	message := payload{}

	for {
		select {
			case message = <- rp:
				feedWindow = append(feedWindow ,message)
			case <- ticker.C:
				process(serialMap, workWindow, feedWindow)
				workWindow = feedWindow		// change feed to work
				feedWindow = []payload{}	// re-init feed
		}
	}
}

func process(serialMap map[int64]int64, workWindow []payload, feedWindow []payload) {
	var pkts,drops,dups,reords int

	var maxRtt, minRtt, totRtt time.Duration
	minRtt = time.Duration(10*time.Second)

	for i,message := range workWindow {
		updateRtt(message, &maxRtt, &minRtt, &totRtt)

		_, ok := serialMap[message.Id]
		if !ok {	// initial packet from this sender ID
			serialMap[message.Id] = message.Serial+1
			pkts++
			continue
		}
		if message.Serial == serialMap[message.Id] {	// correct order
			pkts++
			dups = dups + findPacket(serialMap, workWindow, feedWindow, i+1, message.Id)	// find duplicates
			serialMap[message.Id]++
			continue
		}
		if message.Serial < serialMap[message.Id] {		// lower than expected, re-order that already is handled
			continue
		}

		// message.Serial is larger than expected serial.
		// increment til we catch up
		for ; message.Serial > serialMap[message.Id]; {	// serial larger, drop or re-order
			d := findPacket(serialMap, workWindow, feedWindow, i, message.Id)
			if d == 0 {	// packet loss
				drops++
				pkts++
				serialMap[message.Id]++
				continue
			}
			if d == 1 {	// re-order
				reords++
				pkts++
				serialMap[message.Id]++
				continue
			}
			if d > 1 {	// re-order and duplicates
				reords++
				dups = dups+d
				pkts++
				serialMap[message.Id]++
				continue
			}
		}
		serialMap[message.Id]++
	}

	// check that the last packet in workWindow isn't missing by searching
	// for the next serial in feedWindow
	// add code... for each Id...

	// print some stats
	if pkts > 0 {
		fmt.Print("packets: ", pkts)
		fmt.Print(" drops: ", drops)
		fmt.Printf("(%.2f%%) ", float64(drops)/float64(pkts)*100)
		fmt.Print("re-ordered: ", reords)
		fmt.Printf("(%.2f%%) ", float64(reords)/float64(pkts)*100)
		fmt.Print(" duplicates: ", dups)

		avgRtt := totRtt/time.Duration(pkts)
		fastest := minRtt-avgRtt	// time below avg rtt
		slowest := maxRtt-avgRtt	// time above avg rtt
		fmt.Println(" avg rtt:", avgRtt, "fastest:", fastest, "slowest:", slowest)
	}
}

func findPacket(serialMap map[int64]int64, workWindow []payload, feedWindow []payload, pos int, id int64) int {
	var n int	// number of matching packets

	for _,v := range workWindow[pos:] {
		if v.Id == id {
			if v.Serial == serialMap[v.Id] {
				n++
			}
		}
	}
	for _,v := range feedWindow {
		if v.Id == id {
			if v.Serial == serialMap[v.Id] {
				n++
			}
		}
	}
	return n
}

func updateRtt(message payload, maxRtt *time.Duration, minRtt *time.Duration, totRtt *time.Duration) {
		rtt := message.Rts.Sub(message.Cts)

		*totRtt = *totRtt + rtt
		if rtt < *minRtt {
			*minRtt = rtt
		}
		if rtt > *maxRtt {
			*maxRtt = rtt
		}
}

