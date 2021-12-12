package hub_backend

import (
	"github.com/zyjblockchain/hub-backend/common"
	"gorm.io/gorm/clause"

	"github.com/zyjblockchain/hub-backend/schema"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var log = common.NewLog("mirror-hub")

type Wdb struct {
	Db *gorm.DB
}

func NewWdb(dsn string) *Wdb {
	logLevel := logger.Info
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:          logger.Default.LogMode(logLevel), // 日志 level 设置, prod 使用 warn
		CreateBatchSize: 20,                               // 每次批量插入最大数量
	})
	if err != nil {
		panic(err)
	}

	log.Info("connect wdb success")
	return &Wdb{Db: db}
}

func (w *Wdb) Migrate() error {
	// migrate table
	return w.Db.AutoMigrate(
		&schema.ArticleMark{},
	)
}

func (w *Wdb) Close() {
	d, err := w.Db.DB()
	if err == nil {
		d.Close()
	}
}

func (w *Wdb) GetLastArtMarkHeight() (int64, error) {
	record := schema.ArticleMark{}
	err := w.Db.Model(&schema.ArticleMark{}).Order("block_height desc").Limit(1).Scan(&record).Error
	if err == gorm.ErrRecordNotFound {
		return 0, nil
	}
	return record.BlockHeight, err
}

func (w *Wdb) BatchInsertOrUpdateArtMark(amtArr []*schema.ArticleMark) error {
	return w.Db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "original_content_digest"}},
		UpdateAll: true,
	}).Create(&amtArr).Error
}

func (w *Wdb) GetWaitingArtMark() ([]schema.ArticleMark, error) {
	records := make([]schema.ArticleMark, 0, 200)
	err := w.Db.Model(&schema.ArticleMark{}).Where("status = ?", schema.WaitingStatus).Order("block_height asc").Limit(200).Scan(&records).Error
	return records, err
}

func (w *Wdb) UpdateArtMarkStatus(ocd string, status string) error {
	return w.Db.Model(&schema.ArticleMark{}).Where("original_content_digest = ?", ocd).Update("status", status).Error
}
