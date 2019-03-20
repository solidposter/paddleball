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
	"net"
	"time"
)

func client(rp chan<- payload, id int, addr string, key int, rate int, size int) {
	conn, err := net.Dial("udp",addr)
	if err != nil {
		log.Fatal("Failed to open UDP socket:", err)
	}
	go receiver(rp, conn, key)
	sender(id, conn, key, rate, size)
}

func receiver(rp chan<- payload, conn net.Conn, key int) {
	buf := make([]byte, 65536)
	message := payload{}

	for {
		length, err := conn.Read(buf)
		if err != nil {
			fmt.Println("receiver:", err)
			continue
		}
		rts := time.Now()
		message = decode(buf,length)
		if message.Key != int64(key) {
			fmt.Println("receiver bad key", message)
			continue
		}
		message.Rts = rts
//		message.SetRecvTs()
		rp <- message
	}
}

func sender(id int, conn net.Conn, key int, rate int, size int) {
	var buf *bytes.Buffer

	message := newPayload(id, key, size)

	ticker := time.NewTicker( time.Duration(1000000000/rate) * time.Nanosecond)
	for {
		message.Cts = time.Now()
		buf = message.encode()
		_, err := conn.Write(buf.Bytes())
		if err != nil {
			log.Fatal("Write failed: ", err)
		}
		message.Increment()
		<-ticker.C
	}
}
