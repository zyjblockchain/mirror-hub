package hub_backend

import (
	"testing"
	"time"
)

func TestNewWdb(t *testing.T) {
	tt := time.Unix(time.Now().Unix()-7*24*3600, 0).Format("2006-01-02")
	t.Log(tt)
}
