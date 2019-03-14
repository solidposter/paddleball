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
				fmt.Println("processing done")
		}
	}
}


func process() {
	for i,message := range pslice2 {	// process old data
		nser, ok := serMap[message.Id]
		if ok {
			if message.Serial == nser {
				fmt.Println("correct packet order:", message.Id, nser, message.Serial)
				serMap[message.Id] = message.Serial+1
			} else if message.Serial >  nser { 
				if findPacket(i, message.Id) {
					fmt.Println("packet re-order:", message.Id, nser, message.Serial)
				} else {
					fmt.Println("packet loss:",message.Id, nser)
				}
				serMap[message.Id] = message.Serial+2
			} else {
				continue	// lower than expected serial
			}
		} else {
			serMap[message.Id] = message.Serial+1	// first packet seen for this client ID
		}
	}
}

func findPacket(pos int, id int64) bool {
	for _,v := range pslice2[pos:] {
		if v.Id == id {
			if v.Serial == serMap[v.Id] {
				return true	// packet found, re-order
			}
		}
	}

	for _,v := range pslice1 {	// search the next slice also
		if v.Id == id {
			if v.Serial == serMap[v.Id] {
				return true	// packet found, re-order
			}
		}
	}
	return false	// packet not found, drop
}

