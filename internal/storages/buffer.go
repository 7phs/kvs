package storages

type refCounter interface {
	BufferInUse()
	BufferFree()
}

type Buffer struct {
	refCounter refCounter
	buf        []byte
}

func newBuffer(counter refCounter, buf []byte) Buffer {
	return Buffer{
		refCounter: counter,
		buf:        buf,
	}
}

func (o *Buffer) inUse() {
	o.refCounter.BufferInUse()
}

func (o *Buffer) Free() {
	o.refCounter.BufferFree()

	o.Reset()
}

func (o *Buffer) Copy(data []byte) {
	copy(o.buf, data)
}

func (o *Buffer) Bytes() []byte {
	return o.buf
}

func (o *Buffer) Reset() {
	o.refCounter = nil
	o.buf = nil
}
