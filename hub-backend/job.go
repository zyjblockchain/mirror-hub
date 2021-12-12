package hub_backend

import (
	"github.com/panjf2000/ants/v2"
	"github.com/zyjblockchain/hub-backend/schema"
	"sync"
)

func (h *Hub) runJobs() {
	h.scheduler.Every(10).Second().SingletonMode().Do(h.syncMirrorTxs)
	h.scheduler.Every(1).Minute().SingletonMode().Do(h.syncMirrorArticle)
	h.scheduler.StartAsync()
}

func (h *Hub) syncMirrorTxs() {
	info, err := h.arCli.GetInfo()
	if err != nil {
		log.Error("h.arCli.GetInfo()", "err", err)
		return
	}
	endHeight := info.Height - 3 // stable block

	if endHeight <= h.syncedHeight {
		log.Warn("endHeight can not less than syncedHeight", "syncedHeight", h.syncedHeight, "endHeight", endHeight)
		return
	}
	if endHeight-h.syncedHeight > 50000 {
		endHeight = h.syncedHeight + 50000
	}

	log.Debug("start syncMirrorTxs", "fromBlock", h.syncedHeight, "toBlock", endHeight)
	resTxTags := make([]ResTxTags, 0, 10)
	err = getMirrorTxTags(h.arCli, h.syncedHeight, endHeight, "", &resTxTags)
	if err != nil {
		log.Error("getMirrorTxTags(h.arCli,syncedHeight+1, endHeight,\"\",&resTxTags)", "err", err, "syncedHeight", h.syncedHeight, "endHeight", endHeight)
		return
	}

	amtArr := mergeDigestTags(resTxTags)

	log.Debug("save arTx mark", "number", len(amtArr))
	// insert or update to mysql
	if err := h.wdb.BatchInsertOrUpdateArtMark(amtArr); err != nil {
		log.Error("h.wdb.BatchInsertArtMark(amtArr)", "err", err)
		return
	}
	log.Debug("success syncMirrorTxs", "fromBlock", h.syncedHeight, "toBlock", endHeight)
	h.syncedHeight = endHeight + 1
}

func (h *Hub) syncMirrorArticle() {
	atmArr, err := h.wdb.GetWaitingArtMark()
	if err != nil {
		log.Error("h.wdb.GetWaitingArtMark()", "err", err)
		return
	}
	if len(atmArr) == 0 {
		log.Debug("not need syncMirrorArticle")
		return
	}

	var wg sync.WaitGroup

	p, _ := ants.NewPoolWithFunc(100, func(i interface{}) {
		defer wg.Done()
		atm, ok := i.(schema.ArticleMark)
		if !ok {
			log.Error("i.(schema.ArticleMark) failed", "i", i)
			return
		}
		if err := processMirrorArticle(h.arCli, atm); err != nil {
			log.Error("processMirrorArticle(h.arCli,atm)", "err", err, "atm", atm.ArId)
			// update status failed
			if err := h.wdb.UpdateArtMarkStatus(atm.OriginalContentDigest, schema.FailedStatus); err != nil {
				log.Error("h.wdb.UpdateArtMarkStatus(atm.ID,schema.FailedStatus);", "err", err, "atm", atm.ArId)
			}
			return
		}
		// success update mysql
		if err := h.wdb.UpdateArtMarkStatus(atm.OriginalContentDigest, schema.SuccessStatus); err != nil {
			log.Error("h.wdb.UpdateArtMarkStatus(atm.ID,schema.SuccessStatus);", "err", err, "atm", atm.ArId)
		}
	})
	defer p.Release()

	for _, atm := range atmArr {
		wg.Add(1)
		_ = p.Invoke(atm)
	}
	wg.Wait()
	log.Debug("syncMirrorArticle success", "atm number", len(atmArr))
}
