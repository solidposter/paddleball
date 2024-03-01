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

	"github.com/goccy/go-json"
)

type client struct {
	id int
}

func newclient(id int) *client {
	return &client{
		id: id,
	}
}

func (c *client) probe(addr string, key int) (lport, hport int) {
	req := newPayload(c.id, key, 100)
	nbuf := make([]byte, 1500)

	conn, err := net.Dial("udp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	buffer := new(bytes.Buffer)
	enc := json.NewEncoder(buffer)
	err = enc.Encode(req)
	if err != nil {
		log.Panic(err)
	}

	success := false // set to true on valid response
	for {
		_, err = conn.Write(buffer.Bytes())
		if err != nil {
			log.Fatal(err)
		}
		conn.SetReadDeadline((time.Now().Add(1000 * time.Millisecond)))
		for {
			length, err := conn.Read(nbuf)
			if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
				fmt.Print(".")
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			dec := json.NewDecoder(bytes.NewBuffer(nbuf[:length]))
			err = dec.Decode(&req)
			if err != nil {
				log.Print(err, addr)
				continue
			}
			if req.Key != int64(key) {
				log.Printf("Invalid key %v from %v\n", req.Key, conn.RemoteAddr().String())
				continue
			}
			success = true
			break
		}
		if success {
			break
		}
	}
	return req.Lport, req.Hport
}

func (c *client) start(rp chan<- payload, targetIP string, targetPort string, key int, rate int, size int) {
	conn, err := net.Dial("udp", targetIP+":"+targetPort)
	if err != nil {
		log.Fatal(err)
	}
	go receiver(rp, conn, key)
	sender(c.id, conn, key, rate, size)
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
			log.Printf("Invalid key %v from %v\n", resp.Key, conn.RemoteAddr().String())
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
