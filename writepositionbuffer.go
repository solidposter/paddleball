package main

import (
	"errors"
)

type WritePositionBuffer struct {
	Data     []byte
	WritePos int
	ReadPos  int
}

func NewWritePositionBuffer(bufferSize int) *WritePositionBuffer {
	b := WritePositionBuffer{
		Data:     make([]byte, bufferSize),
		WritePos: 0,
		ReadPos:  0,
	}
	return &b
}

func (w *WritePositionBuffer) Write(p []byte) (n int, err error) {
	copy(w.Data[w.WritePos:], p)
	w.WritePos += len(p)
	return len(p), nil
}

func (w *WritePositionBuffer) Read(p []byte) (n int, err error) {
	copyLength := w.WritePos - w.ReadPos
	if len(p) < copyLength {
		copyLength = len(p)
	}
	copy(p, w.Data[w.ReadPos:w.ReadPos+copyLength])
	w.ReadPos += copyLength
	return copyLength, nil
}

func (w *WritePositionBuffer) ReadByte() (byte, error) {
	if w.ReadPos < w.WritePos {
		b := w.Data[w.ReadPos]
		w.ReadPos += 1
		return b, nil
	} else {
		return 'a', errors.New("Already at the end of reading buffer")
	}
}

func (w *WritePositionBuffer) Reset() {
	w.WritePos = 0
	w.ReadPos = 0
}
