package model

import (
	"crypto/rand"
	"encoding/hex"
	"sync/atomic"
	"time"
)

var idCounter int64

// genID produces a process-unique identifier with the given prefix.
func genID(prefix string) string {
	buf := make([]byte, 5)
	_, _ = rand.Read(buf)
	n := atomic.AddInt64(&idCounter, 1)
	return prefix + hex.EncodeToString(buf) + "-" + time.Now().Format("060102150405") + "-" + itoa(n)
}

func itoa(v int64) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var b [20]byte
	i := len(b)
	for v > 0 {
		i--
		b[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
