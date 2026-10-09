package models

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Fucai3DRecord struct {
	Issue        string `gorm:"column:issue;type:varchar(20);primaryKey;not null;comment:开奖期号" json:"issue"`
	DrawDate     Date   `gorm:"column:draw_date;type:date;not null;comment:开奖日期" json:"draw_date"`
	Num1         int    `gorm:"column:num_1;type:tinyint;not null;comment:开奖号-百位" json:"num_1"`
	Num2         int    `gorm:"column:num_2;type:tinyint;not null;comment:开奖号-十位" json:"num_2"`
	Num3         int    `gorm:"column:num_3;type:tinyint;not null;comment:开奖号-个位" json:"num_3"`
	TestNum1     int    `gorm:"column:test_num_1;type:tinyint;not null;comment:试机号-百位" json:"test_num_1"`
	TestNum2     int    `gorm:"column:test_num_2;type:tinyint;not null;comment:试机号-十位" json:"test_num_2"`
	TestNum3     int    `gorm:"column:test_num_3;type:tinyint;not null;comment:试机号-个位" json:"test_num_3"`
	Machine      int    `gorm:"column:machine;type:tinyint;not null;comment:机型" json:"machine"`
	BallSet      int    `gorm:"column:ball_set;type:tinyint;not null;comment:球号" json:"ball_set"`
	TotalSales   int64  `gorm:"column:total_sales;type:bigint;not null;default:0;comment:投注总额" json:"total_sales"`
	DirectCount  int32  `gorm:"column:direct_count;type:int;not null;default:0;comment:单选注数" json:"direct_count"`
	DirectAmount int64  `gorm:"column:direct_amount;type:bigint;not null;default:0;comment:单选单注金额/总额" json:"direct_amount"`
	Group3Count  int32  `gorm:"column:group3_count;type:int;not null;default:0;comment:组三注数" json:"group3_count"`
	Group3Amount int64  `gorm:"column:group3_amount;type:bigint;not null;default:0;comment:组三单注金额/总额" json:"group3_amount"`
	Group6Count  int32  `gorm:"column:group6_count;type:int;not null;default:0;comment:组六注数" json:"group6_count"`
	Group6Amount int64  `gorm:"column:group6_amount;type:bigint;not null;default:0;comment:组六单注金额/总额" json:"group6_amount"`
}

func (Fucai3DRecord) TableName() string { return "fucai3d_records" }

type Fucai3DRecordStore struct{ DB *gorm.DB }

func (s *Fucai3DRecordStore) Upsert(ctx context.Context, record *Fucai3DRecord) error {
	return s.DB.WithContext(ctx).Clauses(fucai3DConflict()).Create(record).Error
}

func (s *Fucai3DRecordStore) UpsertHistory(ctx context.Context, records []Fucai3DRecord) error {
	if len(records) == 0 {
		return nil
	}
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Clauses(fucai3DConflict()).CreateInBatches(records, 200).Error
	})
}

func fucai3DConflict() clause.OnConflict {
	return clause.OnConflict{Columns: []clause.Column{{Name: "issue"}}, DoUpdates: clause.AssignmentColumns([]string{
		"draw_date", "num_1", "num_2", "num_3", "test_num_1", "test_num_2", "test_num_3", "machine", "ball_set", "total_sales", "direct_count", "direct_amount", "group3_count", "group3_amount", "group6_count", "group6_amount",
	})}
}
