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

func statsengine(rp <-chan payload, rate int, numclients int) {
	ticker := time.NewTicker(time.Second)
	message := payload{}

	for {
		select {
			case message = <- rp:
				pslice1 = append(pslice1,message)
			case <- ticker.C:
				for i,v := range pslice2 {	// process old data
					fmt.Println(i,v)
				}
				pslice2 = pslice1	// copy data
				pslice1 = []payload{}	// zap slice
				fmt.Println("processing done")
		}
	}
}

