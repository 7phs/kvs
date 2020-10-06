package storages

import "time"

type record struct {
	value      Buffer
	expiration time.Time
}

func newRecord(value Buffer, expiration time.Time) record {
	return record{
		value:      value,
		expiration: expiration,
	}
}

func (o *record) get() Buffer {
	o.value.inUse()

	return o.value
}

func (o *record) isExpired() bool {
	return !o.expiration.After(time.Now())
}

func (o *record) reset() {
	o.value.Reset()
}
