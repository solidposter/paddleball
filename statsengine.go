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
var badname map[int64]payload

func statsengine(rp <-chan payload, rate int, numclients int) {
	badname = make(map[int64]payload)

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
	for _,v := range pslice2 {	// process old data
		elem, ok := badname[v.Id]
		if ok {
			if v.Serial == elem.Serial+1 {
				fmt.Println("correct packet order:", v.Id, elem.Serial, v.Serial)
			} else { 
				fmt.Println("wrong packet order:", v.Id, elem.Serial, v.Serial)
			}
		}
		badname[v.Id] = v
	}
}

