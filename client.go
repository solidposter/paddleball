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
	"bytes"
	"encoding/json"
	"log"
	"net"
	"time"
)

func client(rp chan<- payload, id int, addr string, key int, rate int, size int) {
	conn, err := net.Dial("udp", addr)
	if err != nil {
		log.Fatal("client:", err)
	}
	go receiver(rp, conn, key)
	sender(id, conn, key, rate, size)
}

func receiver(rp chan<- payload, conn net.Conn, key int) {
	nbuf := make([]byte, 65536)
	resp := payload{}
	pbdrop := 0 // drop counter

	for {
		length, err := conn.Read(nbuf)
		if err != nil {
			log.Print(err)
			continue
		}
		rts := time.Now() // receive timestamp

		dec := json.NewDecoder(bytes.NewBuffer(nbuf[:length]))
		err = dec.Decode(&resp)
		if err != nil {
			log.Print(err)
			continue
		}
		if resp.Key != int64(key) {
			log.Print("Invalid key", conn.RemoteAddr().String())
			continue
		}

		resp.Rts = rts
		resp.Pbdrop = int64(pbdrop) // copy the drop counter to the packet
		select {
		case rp <- resp: // put the packet in the channel
			pbdrop = 0 // reset the drop counter

		default: // channel full, discard packet, increment drop counter
			pbdrop++
		}
	}
}

func sender(id int, conn net.Conn, key int, rate int, size int) {
	req := newPayload(id, key, size)
	ticker := time.NewTicker(time.Duration(1000000000/rate) * time.Nanosecond)
	for {
		req.Cts = <-ticker.C
		buffer := new(bytes.Buffer)
		enc := json.NewEncoder(buffer)
		err := enc.Encode(req)
		if err != nil {
			log.Panic(err)
		}

		_, err = conn.Write(buffer.Bytes())
		if err != nil {
			log.Print(err)
		}
		req.Serial++
	}
}
