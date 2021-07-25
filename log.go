package wikithing

import (
	"time"

	"git.lan/wikithing/etc/sid"
)

type LogFile struct {
	Entries []LogEntry
}

type LogEntry struct {
	Actor  sid.ID
	Action uint
	Reason string
	When   time.Time
}

// Log action types
const (
	LogActionUnspecified = 0
	LogActionCreate      = 1
	LogActionEdit        = 2
)

// LogActionNames .
var LogActionNames = map[uint]string{
	LogActionUnspecified: "Unspecified Action (an error probably)",
	LogActionCreate:      "Create",
	LogActionEdit:        "Edit",
}
