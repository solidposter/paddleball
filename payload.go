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
	"encoding/gob"
	"log"
	"math/rand"
	"time"
)

type payload struct {
	Id	int64
        Key     int64
        Serial      int64
        Cts     time.Time	// client timestamp
        Sts     time.Time	// server timestamp
        Rts     time.Time	// receiver timestamp
        Data    []byte		// random data
}

func newPayload(id int, key int) payload {
	m :=payload{}
	m.Id = int64(id)	// client id - identify client thread
	m.Key = int64(key)
	m.Data = make([]byte, 32)
	rand.Read(m.Data)
	return m
}

func decode(buffer []byte, length int) payload {
	m := payload{}

	dec := gob.NewDecoder(bytes.NewBuffer(buffer[:length]))
	err := dec.Decode(&m)
	if err != nil {
		log.Fatal("decode error:", err)
	}
	return m
}

func (m payload) encode() *bytes.Buffer {
	var buffer bytes.Buffer

	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(m)
	if err != nil {
		log.Fatal("encode failed:", err)
	}
	return &buffer
}


func (m *payload) GetCts() time.Time {
	return m.Cts
}

func (m *payload) GetKey() int64 {
	return m.Key
}

func (m *payload) Getserial() int64 {
	return m.Serial
}

func (m *payload) GetSts() time.Time {
	return m.Sts
}

func (m *payload) Increment() {
	m.Serial++
}

func (m *payload) SetClientTs() {
	m.Cts = time.Now()
}

func (m *payload) SetRecvTs() {
	m.Rts = time.Now()
}

func (m *payload) SetServerTs() {
	m.Sts = time.Now()
}

