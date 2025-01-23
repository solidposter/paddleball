package main

import (
	"github.com/google/go-cmp/cmp"
	"github.com/goccy/go-json"
	"encoding/gob"
	"crypto/rand"
	"bytes"
	"testing"
	"time"
)

type legacyPayload struct {
	Id     int64     // client ID
	Key    int64     // server key
	Serial int64     // packet serial number
	Lport  int       // Server base port
	Hport  int       // Server highest port
	Cts    time.Time // client timestamp
	Sts    time.Time // server timestamp
	Rts    time.Time // receiver timestamp
	Pbdrop int64     // Paddleball drops
	Data   []byte    // random data
}

func newLegacyPayload(id int, key int, size int) legacyPayload {
	m := legacyPayload{}
	m.Id = int64(id)
	m.Key = int64(key)
	m.Data = make([]byte, size)
	rand.Read(m.Data)
	return m
}

func TestPayloadMarshal(t *testing.T) {
	p := newPayload(1337, 8, 384)
	p.Serial = 123990
	p.Lport = 321
	p.Hport = 65500

	targetP := payload{}

	w := NewWritePositionBuffer(65535)
	for i := 0; i < 50; i++ {
		// Test multiple time to see if the NewWritePositionBuffer reset works properly
		p.Serial += int64(i) * 1000
		p.Lport--
		p.Cts = time.Now().UnixNano()
		p.Sts = time.Now().UnixNano() + 50000
		p.Rts = time.Now().UnixNano() - 90000
		p.MarshalPayload(w)
		w.UnmarshalPayload(&targetP)
		w.Reset()
		if !cmp.Equal(targetP, p) {
			t.Errorf("Wrong object received, got=%s", cmp.Diff(targetP, p))
		}
	}
}

func BenchmarkBinaryMarshal(b *testing.B) {
	p := newPayload(1337, 8, 384)
	p.Serial = 123990
	p.Lport = 321
	p.Hport = 65500
	p.Cts = time.Now().UnixNano()
	p.Sts = time.Now().UnixNano() + 50000
	p.Rts = time.Now().UnixNano() - 90000
	targetP := payload{}

	w := NewWritePositionBuffer(65535)
	for i := 0; i < b.N; i++ {
		p.MarshalPayload(w)
		w.UnmarshalPayload(&targetP)
		w.Reset()
	}
}

func BenchmarkJSONMarshal(b *testing.B) {
	// Benchmarking like previous paddleball did it (new encoder for every package)
	p := newLegacyPayload(1337, 8, 384)
	p.Serial = 123990
	p.Lport = 321
	p.Hport = 65500
	p.Cts = time.Now()
	p.Sts = time.Now()
	p.Rts = time.Now()
	targetP := legacyPayload{}
	nbuf := make([]byte, 0, 65536)

	for i := 0; i < b.N; i++ {
		buffer := bytes.NewBuffer(nbuf)
		enc := json.NewEncoder(buffer)
		_ = enc.Encode(p)

		dec := json.NewDecoder(buffer)
		_ = dec.Decode(&targetP)
		buffer.Reset()
	}
}

func BenchmarkGobMarshal(b *testing.B) {
	// This can't be used for paddleball, since GOB is intended for stateful streaming, but paddleball is stateless udp
	p := newLegacyPayload(1337, 8, 384)
	p.Serial = 123990
	p.Lport = 321
	p.Hport = 65500
	p.Cts = time.Now()
	p.Sts = time.Now()
	p.Rts = time.Now()
	targetP := legacyPayload{}
	nbuf := make([]byte, 0, 65536)

	buffer := bytes.NewBuffer(nbuf)
	enc := gob.NewEncoder(buffer)
	dec := gob.NewDecoder(buffer)
	for i := 0; i < b.N; i++ {
		_ = enc.Encode(p)
		_ = dec.Decode(&targetP)
		buffer.Reset()
	}
}

func BenchmarkGobStatelessMarshal(b *testing.B) {
	// GOB while creating new encoder and decoder for each packet, like it would be needed for paddleball
	p := newLegacyPayload(1337, 8, 384)
	p.Serial = 123990
	p.Lport = 321
	p.Hport = 65500
	p.Cts = time.Now()
	p.Sts = time.Now()
	p.Rts = time.Now()
	targetP := legacyPayload{}
	nbuf := make([]byte, 0, 65536)

	for i := 0; i < b.N; i++ {
		buffer := bytes.NewBuffer(nbuf)
		enc := gob.NewEncoder(buffer)
		_ = enc.Encode(p)

		dec := gob.NewDecoder(buffer)
		_ = dec.Decode(&targetP)
		buffer.Reset()
	}
}
