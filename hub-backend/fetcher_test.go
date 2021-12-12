package hub_backend

import (
	"github.com/everFinance/goar"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_getMirrorTxTags(t *testing.T) {
	arCli := goar.NewClient("https://arweave.net", "http://127.0.0.1:8001")
	resTxTags := make([]ResTxTags, 0, 10)
	err := getMirrorTxTags(arCli, 571604, 629284, "", &resTxTags)
	assert.NoError(t, err)
	t.Log(len(resTxTags))
}
