package commonlog

import (
	"fmt"
	"time"
)

// Event describes a log entry in the common format
type Event struct {
	Host    string
	RFC931  string
	User    string
	Date    time.Time
	Request string
	Status  int
	Bytes   int
	Section string
}

func (e Event) String() string {
	return fmt.Sprintf("host:%v|rfc931:%v|user:%v|date:%v|request:%v|status:%v|bytes:%v",
		e.Host, e.RFC931, e.User, e.Date, e.Request, e.Status, e.Bytes)
}
