package hub_backend

import (
	"encoding/json"
	"fmt"
	"github.com/everFinance/goar"
	"github.com/everFinance/goar/types"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
	"github.com/zyjblockchain/hub-backend/schema"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

type ResTxTags struct {
	ArId         string
	OwnerAddress string
	Tags         []types.Tag
	BlockHeight  int64
}

func getMirrorTxTags(gqlCli *goar.Client, minBlock, maxBlock int64, after string, resTxs *[]ResTxTags) error {
	gql := `query {
					transactions(
						owners:["Ky1c1Kkt-jZ9sY1hvLF5nCf6WWdBhIU5Un_BMYh-t3c"],
						block:{min:` + fmt.Sprintf("%d", minBlock) + `, max:` + fmt.Sprintf("%d", maxBlock) + `},
						first:100,
						after:` + "\"" + after + "\"" + `,
						tags:[
							{
							  name:"Content-Type",
							  values:["application/json"],
							},
							{
							  name:"App-Name",
							  values:["MirrorXYZ"],
							}
						  ]) {
							pageInfo {
								hasNextPage
									  }
							edges {
								node {
								    id
									owner {
										address
									}
									tags{
										name
										value
									  }
									block {
										height
									  }
									}
								cursor
								}
						  }
					}`

	data, err := gqlCli.GraphQL(gql)
	if err != nil {
		return err
	}

	res := struct {
		Transactions struct {
			PageInfo struct {
				HashNextPage bool `json:"hasNextPage"`
			} `json:"pageInfo"`
			Edges []struct {
				Node struct {
					Id    string
					Owner struct{ Address string }
					Tags  []types.Tag
					Block struct{ Height int64 }
				}
				Cursor string
			}
		}
	}{}

	if err = json.Unmarshal(data, &res); err != nil {
		return err
	}

	for _, edge := range res.Transactions.Edges {
		*resTxs = append(*resTxs, ResTxTags{
			ArId:         edge.Node.Id,
			OwnerAddress: edge.Node.Owner.Address,
			Tags:         edge.Node.Tags,
			BlockHeight:  edge.Node.Block.Height,
		})
	}
	// continue search
	if res.Transactions.PageInfo.HashNextPage {
		after = res.Transactions.Edges[len(res.Transactions.Edges)-1].Cursor
		err = getMirrorTxTags(gqlCli, minBlock, maxBlock, after, resTxs)
		return err
	}
	return nil
}

func mergeDigestTags(resTxTags []ResTxTags) schema.ArticleMarkSlice {
	// filter to last Original-Content-Digest
	mmp := make(map[string]ResTxTags, 0) // key: Original-Content-Digest

	for _, res := range resTxTags {
		tagMap := tagsToMap(res.Tags)
		key := tagMap["Original-Content-Digest"]
		// some old mirror tx has not Original-Content-Digest, so key == uuid
		if key == "" {
			key = uuid.NewString()
		}
		mmp[key] = res
	}

	atmArr := make(schema.ArticleMarkSlice, 0, len(mmp))
	for _, res := range mmp {
		tagMap := tagsToMap(res.Tags)
		atm := &schema.ArticleMark{
			ArId:                  res.ArId,
			Contributor:           tagMap["Contributor"],
			OriginalContentDigest: tagMap["Original-Content-Digest"],
			EndContentDigest:      tagMap["Content-Digest"],
			Owner:                 res.OwnerAddress,
			BlockHeight:           res.BlockHeight,
			Status:                schema.WaitingStatus,
		}
		atmArr = append(atmArr, atm)
	}

	// sort by blockHeight
	sort.Sort(atmArr)
	return atmArr
}

func tagsToMap(tags []types.Tag) map[string]string {
	mmp := make(map[string]string, len(tags))
	for _, tag := range tags {
		mmp[tag.Name] = tag.Value
	}
	return mmp
}

func processMirrorArticle(arCli *goar.Client, atm schema.ArticleMark) error {
	data, err := arCli.GetTransactionDataByGateway(atm.ArId)
	// if err != nil {
	// 	log.Error("arCli.GetTransactionDataByGateway(atm.ArId)","err",err,"arId",atm.ArId)
	// 	data, err = arCli.GetTxDataFromPeers(atm.ArId)
	// }
	if err != nil {
		log.Error("get arTx data failed", "arId", atm.ArId)
		return err
	}
	return storeMirrorArt(data, atm)
}

func storeMirrorArt(txData []byte, atm schema.ArticleMark) error {
	content := gjson.ParseBytes(txData).Get("content")
	body := content.Get("body").String()
	timestamp := content.Get("timestamp").Int()
	date := time.Unix(timestamp, 0).Format("2006-01-02T15:04:05Z")

	title := content.Get("title").String()
	title = strings.ReplaceAll(title, "\"", "\\\"")
	title = strings.ReplaceAll(title, "\\'", "\\\"")

	// author := gjson.ParseBytes(txData).Get("authorship.contributor").String()

	ss := strings.ReplaceAll(body, `\`, "")
	// add header
	header := `---
title: ` + "\"" + title + "\"" + `
date: ` + date + `
draft: false
---` + "\n\n" + fmt.Sprintf("###### Author: [%s](https://mirror.xyz/%s)\n\n", atm.Contributor, atm.Contributor)
	ss = header + ss

	// add bottom
	bottom01 := fmt.Sprintf("###### [arweave tx: %s](https://viewblock.io/arweave/tx/%s)\n", atm.ArId, atm.ArId)
	bottom02 := fmt.Sprintf("###### [ethereum address: %s](https://etherscan.io/address/%s)\n", atm.Contributor, atm.Contributor)
	bottom03 := fmt.Sprintf("###### content digest: %s\n", atm.EndContentDigest)
	ss = ss + "\n" + "---\n" + bottom01 + bottom02 + bottom03 + "---"

	basePath := fmt.Sprintf("../mirror-hub.com/content/post/%s/", atm.Contributor)
	if err := os.MkdirAll(basePath, 0777); err != nil {
		log.Error("os.MkdirAll(basePath,0777)", "err", err, "basePath", basePath)
		return err
	}
	fileName := fmt.Sprintf("%s.md", atm.OriginalContentDigest)
	if err := ioutil.WriteFile(path.Join(basePath, fileName), []byte(ss), 0777); err != nil {
		return err
	}
	return nil
}
