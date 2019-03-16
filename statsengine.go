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

type engineInfo struct {
	numClients, rate int
	drops, dups, reords, totPkts int64
	minRtt, maxRtt, totRtt time.Duration
}

func statsEngine(rp <-chan payload, gei *engineInfo) {
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
				process(workWindow, feedWindow, serialMap, gei)
				workWindow = feedWindow		// change feed to work
				feedWindow = []payload{}	// re-init feed
		}
	}
}

func process(workWindow []payload, feedWindow []payload, serialMap map[int64]int64, gei *engineInfo) {
	lei := engineInfo {}			// local engine info
	lei.minRtt = time.Duration(1*time.Hour)	// minRtt must not be zero
	lei.rate = gei.rate
	lei.numClients = gei.numClients

	for i,message := range workWindow {
		updateRtt(message, &lei)

		_, ok := serialMap[message.Id]
		if !ok {	// initial packet from this sender ID
			serialMap[message.Id] = message.Serial+1
			lei.totPkts++
			continue
		}
		if message.Serial == serialMap[message.Id] {	// correct order
			lei.totPkts++
			lei.dups = lei.dups + findPacket(serialMap, workWindow, feedWindow, i+1, message.Id)	// find duplicates
			serialMap[message.Id]++
			continue
		}
		if message.Serial < serialMap[message.Id] {		// lower than expected, re-order that already is handled
			lei.totPkts++
			continue
		}

		// message.Serial is larger than expected serial.
		// increment til we catch up
		for ; message.Serial > serialMap[message.Id]; {	// serial larger, drop or re-order
			d := findPacket(serialMap, workWindow, feedWindow, i, message.Id)
			if d == 0 {	// packet loss
				lei.drops++
				lei.totPkts++
				serialMap[message.Id]++
				continue
			}
			if d == 1 {	// re-order
				lei.reords++
				lei.totPkts++
				serialMap[message.Id]++
				continue
			}
			if d > 1 {	// re-order and duplicates
				lei.reords++
				lei.dups = lei.dups+d
				lei.totPkts++
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
	if lei.totPkts > 0 {
		fmt.Print("packets: ", lei.totPkts)
		fmt.Print(" drops: ", lei.drops)
		fmt.Printf("(%.2f%%) ", float64(lei.drops)/float64(lei.totPkts)*100)
		fmt.Print("re-ordered: ", lei.reords)
		fmt.Printf("(%.2f%%) ", float64(lei.reords)/float64(lei.totPkts)*100)
		fmt.Print(" duplicates: ", lei.dups)

		avgRtt := lei.totRtt/time.Duration(lei.totPkts)
		fastest := lei.minRtt-avgRtt	// time below avg rtt
		slowest := lei.maxRtt-avgRtt	// time above avg rtt
		fmt.Print(" avg rtt: ", avgRtt, " fastest: ", fastest, " slowest: +", slowest)
		fmt.Println()
	}
	// update the global stats
	gei.dups = gei.dups + lei.dups
	gei.reords = gei.reords + lei.reords
	gei.totPkts = gei.totPkts + lei.totPkts
	gei.totRtt = gei.totRtt + lei.totRtt
	if gei.minRtt > lei.minRtt {
		gei.minRtt = lei.minRtt
	}
	if gei.maxRtt < lei.maxRtt {
		gei.maxRtt = lei.maxRtt
	}

}

func findPacket(serialMap map[int64]int64, workWindow []payload, feedWindow []payload, pos int, id int64) int64 {
	var n int64	// number of matching packets

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

func updateRtt(message payload, lei *engineInfo) {
		rtt := message.Rts.Sub(message.Cts)

		lei.totRtt = lei.totRtt + rtt
		if rtt < lei.minRtt {
			lei.minRtt = rtt
		}
		if rtt > lei.maxRtt {
			lei.maxRtt = rtt
		}
}

