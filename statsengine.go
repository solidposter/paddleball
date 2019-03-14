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

var pslice1 = []payload{}	// live data slice, data is fed here
var pslice2 = []payload{}	// old data slice, analysis is done here
var serMap map[int64]int64	// expected serial number, key = client ID

func statsengine(rp <-chan payload, rate int, numclients int) {
	serMap = make(map[int64]int64)

	ticker := time.NewTicker(time.Second)
	message := payload{}

	for {
		select {
			case message = <- rp:
				pslice1 = append(pslice1,message)
			case <- ticker.C:
				process()
				pslice2 = pslice1	// copy data
				pslice1 = []payload{}	// zap slice
		}
	}
}

func process() {
	var pkts,drops,dups,reords int

	for i,message := range pslice2 {
		nser, ok := serMap[message.Id]
		if ok {
			if message.Serial == nser {	// correct order
				pkts++
				dups = dups + findPacket(i+1, message.Id)	// find duplicates
				serMap[message.Id] = message.Serial+1
			} else if message.Serial >  nser {	// serial larger, drop or re-order
				d := findPacket(i, message.Id)
				if d == 0 {
					drops++
					pkts++	// increment packet counter for lost packet
					serMap[message.Id] = message.Serial+2	// lost packet
					fmt.Println("packet loss:",message.Id, nser, message.Serial)
				} else {
					reords++
					pkts++
					dups = dups+d-1	// if we find one here it isn't a duplicate
					serMap[message.Id] = message.Serial+1
					fmt.Println("packet re-order:", message.Id, nser, message.Serial)
				}
			} else {	// lower than expected serial, re-order that already was handled.
				continue
			}
		} else {	// first packet seen for this client ID
			serMap[message.Id] = message.Serial+1
		}
	}

	// check that the last packet in pslice1 isn't missing by searching
	// for the next serial in pclice2
	// add code... for each Id...


	if pkts > 0 {
		fmt.Print("packets: ", pkts)
		fmt.Print(" drops: ", drops)
		fmt.Printf("(%.2f%%) ", float64(drops)/float64(pkts)*100)
		fmt.Print(" re-ordered: ", reords)
		fmt.Printf("(%.2f%%) ", float64(reords)/float64(pkts)*100)
		fmt.Println(" duplicates:", dups)
	}
}

func findPacket(pos int, id int64) int {
	var n int	// number of matching packets

	for _,v := range pslice2[pos:] {
		if v.Id == id {
			if v.Serial == serMap[v.Id] {
				n++
			}
		}
	}
	for _,v := range pslice1 {
		if v.Id == id {
			if v.Serial == serMap[v.Id] {
				n++
			}
		}
	}
	return n
}

