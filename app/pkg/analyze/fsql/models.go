// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package fsql

import (
	"time"
)

type Session struct {
	Contextid         string
	Starttime         time.Time
	Endtime           time.Time
	Selectedid        string
	Selectedkind      string
	TotalInputTokens  int64
	TotalOutputTokens int64
	Proto             []byte
}
