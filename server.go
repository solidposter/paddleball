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
	"fmt"
	"log"
	"math/rand"
	"net"
)

func server(port string, key int) {
	var ebuf *bytes.Buffer
	nbuf := make([]byte, 65536)

	serverkey := int64(key)
	if serverkey == 0 {
		serverkey = rand.Int63()
	}

	fmt.Print("Starting server mode, ")
	pc, err := net.ListenPacket("udp", "0.0.0.0:"+port)
	if err != nil {
		log.Fatal("server:", err)
	}
	fmt.Println("listening on", pc.LocalAddr(), "with server key", serverkey)

	for {
		length, addr, err := pc.ReadFrom(nbuf)
		if err != nil {
			continue
		}

		message := decode(nbuf, length)
		if message.Key != serverkey {
			continue
		}

		message.SetServerTs()
		ebuf = message.encode()
		pc.WriteTo(ebuf.Bytes(), addr)
	}
}
