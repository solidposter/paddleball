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
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type payload struct {
	Id     int64     // client ID
	Key    int64     // server key
	Serial int64     // packet serial number
	Cts    time.Time // client timestamp
	Sts    time.Time // server timestamp
	Rts    time.Time // receiver timestamp
	Pbdrop int64     // Paddleball drops
	Data   []byte    // random data
}

func newPayload(id int, key int, size int) payload {
	m := payload{}
	m.Id = int64(id)
	m.Key = int64(key)
	m.Data = make([]byte, size)
	rand.Read(m.Data)
	return m
}

func decode(buffer []byte, length int) payload {
	m := payload{}

	dec := json.NewDecoder(bytes.NewBuffer(buffer[:length]))
	err := dec.Decode(&m)
	if err != nil {
		fmt.Println("decode error:", err) // do I care ?
	}
	return m
}

func (m payload) encode() *bytes.Buffer {
	var buffer bytes.Buffer

	enc := json.NewEncoder(&buffer)
	err := enc.Encode(m)
	if err != nil {
		log.Fatal("encode failed:", err)
	}
	return &buffer
}
