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

func server(port string, key int64, lport int, hport int) {
	req := payload{}
	nbuf := make([]byte, 65536)

	conn, err := net.ListenPacket("udp", "0.0.0.0:"+port)
	if err != nil {
		log.Fatal("server:", err)
	}
	for {
		length, addr, err := conn.ReadFrom(nbuf)
		if err != nil {
			log.Print(err)
			continue
		}

		dec := json.NewDecoder(bytes.NewBuffer(nbuf[:length]))
		err = dec.Decode(&req)
		if err != nil {
			log.Print(err, addr)
			continue
		}
		if req.Key != key {
			continue
		}

		req.Sts = time.Now()
		req.Lport = lport
		req.Hport = hport
		buffer := new(bytes.Buffer)
		enc := json.NewEncoder(buffer)
		err = enc.Encode(req)
		if err != nil {
			log.Panic(err)
		}

		_, err = conn.WriteTo(buffer.Bytes(), addr)
		if err != nil {
			log.Print(err)
			continue
		}
	}
}
