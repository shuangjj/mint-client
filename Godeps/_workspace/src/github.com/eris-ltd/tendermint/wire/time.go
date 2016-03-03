package wire

import (
	"io"
	"time"
)

/*
Writes nanoseconds since epoch but with millisecond precision.
This is to ease compatibility with Javascript etc.
*/

func WriteTime(t time.Time, w io.Writer, n *int64, err *error) {
	nanosecs := t.UnixNano()
	millisecs := nanosecs / 1000000
	WriteInt64(millisecs*1000000, w, n, err)
}

func ReadTime(r io.Reader, n *int64, err *error) time.Time {
	t := ReadInt64(r, n, err)
	return time.Unix(0, (t/1000000)*1000000)
}
