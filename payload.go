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
	"crypto/rand"
	"encoding/binary"
)

type payload struct {
	Id     int64  // client ID
	Key    int64  // server key
	Serial int64  // packet serial number
	Lport  int64  // Server base port
	Hport  int64  // Server highest port
	Cts    int64  // client timestamp
	Sts    int64  // server timestamp
	Rts    int64  // receiver timestamp
	Pbdrop int64  // Paddleball drops
	Data   []byte // random data
}

func newPayload(id int, key int, size int) payload {
	m := payload{}
	m.Id = int64(id)
	m.Key = int64(key)
	m.Data = make([]byte, size)
	rand.Read(m.Data)
	return m
}

func (p *payload) MarshalPayload(w *WritePositionBuffer) error {
	var len int
	//var err error
	_ = binary.Write(w, binary.LittleEndian, p.Id)
	_ = binary.Write(w, binary.LittleEndian, p.Key)
	_ = binary.Write(w, binary.LittleEndian, p.Serial)
	_ = binary.Write(w, binary.LittleEndian, p.Lport)
	_ = binary.Write(w, binary.LittleEndian, p.Hport)
	_ = binary.Write(w, binary.LittleEndian, p.Cts)
	_ = binary.Write(w, binary.LittleEndian, p.Sts)
	_ = binary.Write(w, binary.LittleEndian, p.Rts)
	_ = binary.Write(w, binary.LittleEndian, p.Pbdrop)
	len = copy(w.Data[w.WritePos:], p.Data)
	w.WritePos += len
	return nil
}

func (w *WritePositionBuffer) UnmarshalPayload(p *payload) error {
	_ = binary.Read(w, binary.LittleEndian, &p.Id)
	_ = binary.Read(w, binary.LittleEndian, &p.Key)
	_ = binary.Read(w, binary.LittleEndian, &p.Serial)
	_ = binary.Read(w, binary.LittleEndian, &p.Lport)
	_ = binary.Read(w, binary.LittleEndian, &p.Hport)
	_ = binary.Read(w, binary.LittleEndian, &p.Cts)
	_ = binary.Read(w, binary.LittleEndian, &p.Sts)
	_ = binary.Read(w, binary.LittleEndian, &p.Rts)
	_ = binary.Read(w, binary.LittleEndian, &p.Pbdrop)
	dataLength := w.WritePos - w.ReadPos
	if dataLength != len(p.Data) {
		p.Data = make([]byte, dataLength)
	}
	copy(p.Data, w.Data[w.ReadPos:w.WritePos])
	return nil
}
