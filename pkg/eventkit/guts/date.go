package guts

import (
	"time"
)

// type for representing date-only times in JSON as yyyy-mm-dd
// (dates without time component)
type Date struct {
	time.Time
}

func (d *Date) MarshalJSON() ([]byte, error) {
	out := d.Format(`"2006-01-02"`)
	return []byte(out), nil
}

func (d *Date) UnmarshalJSON(b []byte) error {
	parsed, err := time.ParseInLocation(`"2006-01-02"`, string(b), time.UTC)
	d.Time = parsed
	return err
}
