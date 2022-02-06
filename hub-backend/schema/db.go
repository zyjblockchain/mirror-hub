package schema

import (
	"time"
)

type ArticleMark struct {
	CreatedAt             time.Time
	UpdatedAt             time.Time
	ArId                  string `gorm:"type:varchar(43);uniqueIndex"`
	Contributor           string `gorm:"index:idx01"`
	OriginalContentDigest string `gorm:"primarykey"`
	EndContentDigest      string
	Owner                 string // mirror arTx sender
	BlockHeight           int64  `gorm:"index:idx03"`
	Timestamp             int64  `gorm:"index:idx05"`
	Status                string `gorm:"index:idx04"` // waiting, update, pending, success, failed
}

func (a ArticleMark) TableName() string {
	return "article_mark_1"
}

type ArticleMarkSlice []*ArticleMark

func (a ArticleMarkSlice) Len() int {
	return len(a)
}

func (a ArticleMarkSlice) Less(i, j int) bool {
	return a[i].BlockHeight < a[j].BlockHeight
}

func (a ArticleMarkSlice) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
