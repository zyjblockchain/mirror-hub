package hub_backend

import (
	"github.com/everFinance/goar"
	"github.com/go-co-op/gocron"
	"time"
)

type Hub struct {
	wdb          *Wdb
	arCli        *goar.Client
	scheduler    *gocron.Scheduler
	syncedHeight int64
}

func New(dsn string, arNode string) *Hub {
	db := NewWdb(dsn)
	if err := db.Migrate(); err != nil {
		panic(err)
	}
	syncedHeight, err := db.GetLastArtMarkHeight()
	if err != nil {
		panic(err)
	}
	if syncedHeight == 0 {
		syncedHeight = 571604
	}
	return &Hub{
		wdb:          db,
		arCli:        goar.NewClient(arNode),
		scheduler:    gocron.NewScheduler(time.UTC),
		syncedHeight: syncedHeight,
	}
}

func (h *Hub) Run() {
	go h.runJobs()
}

func (h *Hub) Close() {
	h.wdb.Close()
}
