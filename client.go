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
	"log/slog"
	"net"
	"os"
	"time"
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

	var conn net.Conn
	var err error
	for {
		conn, err = net.Dial("udp", addr)
		if err != nil {
			slog.Warn("Dial() failed", "error", err)
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}
	defer conn.Close()

	packetBuffer := NewWritePositionBuffer(65535)
	err = req.MarshalPayload(packetBuffer)
	if err != nil {
		slog.Error("Encode() failed", "error", err)
		os.Exit(1)
	}

	success := false // set to true on valid response
	for {
		_, err = conn.Write(packetBuffer.Data[:packetBuffer.WritePos])
		packetBuffer.Reset()
		if err != nil {
			slog.Warn("Write failed", "error", err)
			time.Sleep(1 * time.Second)
			continue
		}
		conn.SetReadDeadline((time.Now().Add(1000 * time.Millisecond)))
		for {
			packetBuffer.WritePos, err = conn.Read(packetBuffer.Data)
			if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
				fmt.Print(".")
				break
			}
			if err != nil {
				slog.Warn("Read() failed", "error", err)
				time.Sleep(1 * time.Second)
				break
			}
			err = packetBuffer.UnmarshalPayload(&req)
			if err != nil {
				slog.Warn("Decode() failed", "error", err)
				continue
			}
			if req.Key != int64(key) {
				slog.Info("Invalid key on probe", "key", req.Key, "host", conn.RemoteAddr().String())
				continue
			}
			success = true
			break
		}
		if success {
			break
		}
	}
	return int(req.Lport), int(req.Hport)
}

func (c *client) start(rp chan<- payload, targetIP string, targetPort string, key int, rate int, size int) {
	conn, err := net.Dial("udp", targetIP+":"+targetPort)
	if err != nil {
		slog.Error("net.Dial() failed", "error", err)
		os.Exit(1)
	}
	go receiver(rp, conn, key)
	sender(c.id, conn, key, rate, size)
}

func receiver(rp chan<- payload, conn net.Conn, key int) {
	resp := payload{}
	pbdrop := 0 // drop counter
	packetBuffer := NewWritePositionBuffer(65535)
	var err error
	var rts time.Time

	for {
		packetBuffer.WritePos, err = conn.Read(packetBuffer.Data)
		if err != nil {
			slog.Warn("Read() failed", "error", err)
			continue
		}
		rts = time.Now() // receive timestamp

		err = packetBuffer.UnmarshalPayload(&resp)
		packetBuffer.Reset()
		if err != nil {
			slog.Warn("Decode() failed", "error", err)
			continue
		}
		if resp.Key != int64(key) {
			slog.Info("Invalid key on receiver", "key", resp.Key, "host", conn.RemoteAddr().String())
			continue
		}

		resp.Rts = rts.UnixNano()
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
	var tickerTime time.Time
	packetBuffer := NewWritePositionBuffer(65535)
	for {
		tickerTime = <-ticker.C
		req.Cts = tickerTime.UnixNano()
		err := req.MarshalPayload(packetBuffer)
		if err != nil {
			slog.Error("Encode() failed", "error", err)
			os.Exit(1)
		}

		_, err = conn.Write(packetBuffer.Data[:packetBuffer.WritePos])
		packetBuffer.Reset()
		if err != nil {
			slog.Warn("Write() failed", "error", err)
		}
		req.Serial++
	}
}
