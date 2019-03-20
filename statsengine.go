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
	Drops, Dups, Reords, TotPkts int64
	MinRtt, MaxRtt, TotRtt int64	// nanoseconds
}

func statsEngine(rp <-chan payload, global *packetStats,  printJson bool) {
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

				if printJson {
					statsPrint(&local, printJson)
				} else {
					statsPrint(&local, printJson)
					fmt.Print(" queue: ",len(rp),"/",cap(rp))
					fmt.Println()
				}
		}
	}
}

func process(workWindow []payload, feedWindow []payload, serialNumbers map[int64]int64) packetStats {
	local := packetStats {}			// local engine info
	local.MinRtt = 1000000000*3600		// MinRtt must not be zero, set 1h in ns

	for position, message := range workWindow {
		updateRtt(message, &local)

		_, ok := serialNumbers[message.Id]
		if !ok {	// initial packet from this sender ID
			serialNumbers[message.Id] = message.Serial+1
			local.TotPkts++
			continue
		}
		if message.Serial == serialNumbers[message.Id] {	// correct order
			local.TotPkts++
			local.Dups = local.Dups + findPacket(serialNumbers, workWindow, feedWindow, position+1, message.Id)	// find duplicates
			serialNumbers[message.Id]++
			continue
		}
		if message.Serial < serialNumbers[message.Id] {		// lower than expected, re-order that already is handled
			local.TotPkts++
			continue
		}

		// message.Serial is larger than expected serial.
		// increment til we catch up
		for ; message.Serial > serialNumbers[message.Id]; {	// serial larger, drop or re-order
			matches := findPacket(serialNumbers, workWindow, feedWindow, position, message.Id)
			if matches == 0 {	// packet loss
				local.Drops++
				local.TotPkts++
				serialNumbers[message.Id]++
				continue
			}
			if matches == 1 {	// re-order
				local.Reords++
				local.TotPkts++
				serialNumbers[message.Id]++
				continue
			}
			if matches > 1 {	// re-order and duplicates
				local.Reords++
				local.Dups = local.Dups+matches
				local.TotPkts++
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

func statsPrint(ei *packetStats, printJson bool) {
	if ei.TotPkts == 0 {
		return
	}

	if printJson {
		b, err := json.Marshal(ei)
		if err != nil {
			fmt.Println("statsPrint error:",err)
		} else {
			os.Stdout.Write(b)
			fmt.Println()
		}
		return
	}

	fmt.Print("packets: ", ei.TotPkts)
	fmt.Print(" Drops: ", ei.Drops)
	fmt.Printf("(%.2f%%) ", float64(ei.Drops)/float64(ei.TotPkts)*100)
	fmt.Print("re-ordered: ", ei.Reords)
	fmt.Printf("(%.2f%%) ", float64(ei.Reords)/float64(ei.TotPkts)*100)
	fmt.Print("duplicates: ", ei.Dups)

	avgRtt := float64(ei.TotRtt/ei.TotPkts)
	fastest := float64(ei.MinRtt)-avgRtt	// time below avg rtt
	slowest := float64(ei.MaxRtt)-avgRtt	// time above avg rtt
	// convert from ns to ms
	avgRtt = avgRtt / 1000000
	fastest = fastest / 1000000
	slowest = slowest / 1000000
	fmt.Print(" avg rtt: ", avgRtt, "ms fastest: ", fastest, "ms slowest: +", slowest,"ms")
}

func statsUpdate(global *packetStats, local packetStats) {
	global.Drops = global.Drops + local.Drops
	global.Dups = global.Dups + local.Dups
	global.Reords = global.Reords + local.Reords
	global.TotPkts = global.TotPkts + local.TotPkts
	global.TotRtt = global.TotRtt + local.TotRtt
	if global.MinRtt > local.MinRtt {
		global.MinRtt = local.MinRtt
	}
	if global.MaxRtt < local.MaxRtt {
		global.MaxRtt = local.MaxRtt
	}
}

func updateRtt(message payload, local *packetStats) {
		rtt := int64( message.Rts.Sub(message.Cts) )

		local.TotRtt = local.TotRtt + rtt
		if rtt < local.MinRtt {
			local.MinRtt = rtt
		}
		if rtt > local.MaxRtt {
			local.MaxRtt = rtt
		}
}
