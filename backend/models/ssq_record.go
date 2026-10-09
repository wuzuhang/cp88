package models

import (
	"context"
	"database/sql/driver"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SSQRecord 与 ssq_records 表一一对应，不增加 ID 或时间戳列。
type SSQRecord struct {
	Issue        string `gorm:"column:issue;type:varchar(20);primaryKey;not null;comment:开奖期号" json:"issue"`
	DrawDate     Date   `gorm:"column:draw_date;type:date;not null;comment:开奖日期" json:"draw_date"`
	Red1         int    `gorm:"column:red_1;type:tinyint;not null;comment:红号1" json:"red_1"`
	Red2         int    `gorm:"column:red_2;type:tinyint;not null;comment:红号2" json:"red_2"`
	Red3         int    `gorm:"column:red_3;type:tinyint;not null;comment:红号3" json:"red_3"`
	Red4         int    `gorm:"column:red_4;type:tinyint;not null;comment:红号4" json:"red_4"`
	Red5         int    `gorm:"column:red_5;type:tinyint;not null;comment:红号5" json:"red_5"`
	Red6         int    `gorm:"column:red_6;type:tinyint;not null;comment:红号6" json:"red_6"`
	Blue         int    `gorm:"column:blue;type:tinyint;not null;comment:蓝号" json:"blue"`
	SeqRed1      int    `gorm:"column:seq_red_1;type:tinyint;not null;comment:红号出球顺序1" json:"seq_red_1"`
	SeqRed2      int    `gorm:"column:seq_red_2;type:tinyint;not null;comment:红号出球顺序2" json:"seq_red_2"`
	SeqRed3      int    `gorm:"column:seq_red_3;type:tinyint;not null;comment:红号出球顺序3" json:"seq_red_3"`
	SeqRed4      int    `gorm:"column:seq_red_4;type:tinyint;not null;comment:红号出球顺序4" json:"seq_red_4"`
	SeqRed5      int    `gorm:"column:seq_red_5;type:tinyint;not null;comment:红号出球顺序5" json:"seq_red_5"`
	SeqRed6      int    `gorm:"column:seq_red_6;type:tinyint;not null;comment:红号出球顺序6" json:"seq_red_6"`
	TotalSales   int64  `gorm:"column:total_sales;type:bigint;not null;default:0;comment:投注总额" json:"total_sales"`
	PoolAmount   int64  `gorm:"column:pool_amount;type:bigint;not null;default:0;comment:奖池金额" json:"pool_amount"`
	Prize1Count  int32  `gorm:"column:prize1_count;type:int;not null;default:0;comment:一等奖注数" json:"prize1_count"`
	Prize1Amount int64  `gorm:"column:prize1_amount;type:bigint;not null;default:0;comment:一等奖金额" json:"prize1_amount"`
	Prize2Count  int32  `gorm:"column:prize2_count;type:int;not null;default:0;comment:二等奖注数" json:"prize2_count"`
	Prize2Amount int64  `gorm:"column:prize2_amount;type:bigint;not null;default:0;comment:二等奖金额" json:"prize2_amount"`
	Prize3Count  int32  `gorm:"column:prize3_count;type:int;not null;default:0;comment:三等奖注数" json:"prize3_count"`
	Prize3Amount int64  `gorm:"column:prize3_amount;type:bigint;not null;default:0;comment:三等奖金额" json:"prize3_amount"`
	Prize4Count  int32  `gorm:"column:prize4_count;type:int;not null;default:0;comment:四等奖注数" json:"prize4_count"`
	Prize4Amount int64  `gorm:"column:prize4_amount;type:bigint;not null;default:0;comment:四等奖金额" json:"prize4_amount"`
	Prize5Count  int32  `gorm:"column:prize5_count;type:int;not null;default:0;comment:五等奖注数" json:"prize5_count"`
	Prize5Amount int64  `gorm:"column:prize5_amount;type:bigint;not null;default:0;comment:五等奖金额" json:"prize5_amount"`
	Prize6Count  int32  `gorm:"column:prize6_count;type:int;not null;default:0;comment:六等奖注数" json:"prize6_count"`
	Prize6Amount int64  `gorm:"column:prize6_amount;type:bigint;not null;default:0;comment:六等奖金额" json:"prize6_amount"`
}

func (SSQRecord) TableName() string { return "ssq_records" }

// Date 按日期字符串写入 MySQL，避免 DSN 时区转换导致日期偏移。
type Date string

func (d Date) Value() (driver.Value, error) { return string(d), nil }

func (d *Date) Scan(value any) error {
	var text string
	switch v := value.(type) {
	case time.Time:
		text = v.Format("2006-01-02")
	case string:
		text = v
	case []byte:
		text = string(v)
	default:
		return fmt.Errorf("不支持的日期类型: %T", value)
	}
	if _, err := time.Parse("2006-01-02", text); err != nil {
		return err
	}
	*d = Date(text)
	return nil
}

type SSQRecordStore struct{ DB *gorm.DB }

// Upsert 重复执行时以期号为唯一键更新完整开奖数据。
func (s *SSQRecordStore) Upsert(ctx context.Context, record *SSQRecord) error {
	return s.DB.WithContext(ctx).Clauses(ssqConflict()).Create(record).Error
}

// UpsertHistory 将所有批次放在一个事务中，失败时整体回滚。
func (s *SSQRecordStore) UpsertHistory(ctx context.Context, records []SSQRecord) error {
	if len(records) == 0 {
		return nil
	}
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Clauses(ssqConflict()).CreateInBatches(records, 200).Error
	})
}

func ssqConflict() clause.OnConflict {
	return clause.OnConflict{
		Columns: []clause.Column{{Name: "issue"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"draw_date", "red_1", "red_2", "red_3", "red_4", "red_5", "red_6", "blue", "seq_red_1", "seq_red_2", "seq_red_3", "seq_red_4", "seq_red_5", "seq_red_6", "total_sales", "pool_amount", "prize1_count", "prize1_amount", "prize2_count", "prize2_amount", "prize3_count", "prize3_amount", "prize4_count", "prize4_amount", "prize5_count", "prize5_amount", "prize6_count", "prize6_amount",
		}),
	}
}
