package main

import (
	"github.com/google/go-cmp/cmp"
	"testing"
	"time"
)

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
