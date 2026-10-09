package service

import (
	"bufio"
	"context"
	"fmt"
	"gorm.io/gorm"
	"io"
	"net/http"
	"shuangseqiu/backend/models"
	"shuangseqiu/utils"
	"strconv"
	"strings"
	"time"
)

type fucai3DWriter interface {
	Upsert(context.Context, *models.Fucai3DRecord) error
}

type Fucai3DSyncService struct {
	downloadService *DrawSyncService
	store           fucai3DWriter
}

func NewFucai3DSyncService(db *gorm.DB, sourceURL string) *Fucai3DSyncService {
	return &Fucai3DSyncService{
		downloadService: &DrawSyncService{client: &http.Client{Timeout: 30 * time.Second}, sourceURL: sourceURL},
		store:           &models.Fucai3DRecordStore{DB: db},
	}
}

func (s *Fucai3DSyncService) FetchHistorySnapshot(ctx context.Context) ([]models.Fucai3DRecord, []byte, error) {
	body, err := s.downloadService.download(ctx)
	if err != nil {
		return nil, nil, err
	}
	records, err := ParseFucai3DHistory(strings.NewReader(string(body)))
	return records, body, err
}

func (s *Fucai3DSyncService) SyncPreviousDay(ctx context.Context, now time.Time) (string, error) {
	local := now.In(utils.ChinaTimezone)
	target := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, utils.ChinaTimezone).AddDate(0, 0, -1).Format("2006-01-02")
	body, err := s.downloadService.download(ctx)
	if err != nil {
		return "", err
	}
	scanner := bufio.NewScanner(strings.NewReader(string(body)))
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	var record *models.Fucai3DRecord
	for scanner.Scan() {
		fields := strings.Fields(strings.TrimPrefix(scanner.Text(), "\uFEFF"))
		if len(fields) < 2 || fields[1] != target {
			continue
		}
		if record != nil {
			return "", fmt.Errorf("日期 %s 出现多条开奖记录", target)
		}
		record, err = parseFucai3DLine(fields)
		if err != nil {
			return "", err
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	if record == nil {
		return "", fmt.Errorf("开奖文件中尚无 %s 的开奖结果", target)
	}
	if err := s.store.Upsert(ctx, record); err != nil {
		return "", fmt.Errorf("保存期号 %s: %w", record.Issue, err)
	}
	return record.Issue, nil
}

func ParseFucai3DHistory(reader io.Reader) ([]models.Fucai3DRecord, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	records := make([]models.Fucai3DRecord, 0, 9000)
	seen := make(map[string]bool)
	line := 0
	for scanner.Scan() {
		line++
		fields := strings.Fields(strings.TrimPrefix(scanner.Text(), "\uFEFF"))
		if len(fields) == 0 {
			continue
		}
		r, err := parseFucai3DLine(fields)
		if err != nil {
			return nil, fmt.Errorf("福彩3D历史数据第 %d 行: %w", line, err)
		}
		if seen[r.Issue] {
			return nil, fmt.Errorf("历史数据重复期号: %s", r.Issue)
		}
		seen[r.Issue] = true
		records = append(records, *r)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("历史数据为空")
	}
	return records, nil
}

func parseFucai3DLine(fields []string) (*models.Fucai3DRecord, error) {
	if len(fields) != 17 {
		return nil, fmt.Errorf("开奖数据应为 17 列，实际 %d 列", len(fields))
	}
	if len(fields[0]) != 7 {
		return nil, fmt.Errorf("期号必须为 7 位数字")
	}
	for _, c := range fields[0] {
		if c < '0' || c > '9' {
			return nil, fmt.Errorf("期号必须为数字")
		}
	}
	date, err := time.ParseInLocation("2006-01-02", fields[1], utils.ChinaTimezone)
	if err != nil {
		return nil, fmt.Errorf("开奖日期格式错误: %w", err)
	}
	n := make([]int64, 15)
	for i := range n {
		value, err := strconv.ParseInt(fields[i+2], 10, 64)
		if err != nil || value < 0 {
			return nil, fmt.Errorf("第 %d 列必须为非负整数", i+3)
		}
		if i < 6 && value > 9 {
			return nil, fmt.Errorf("第 %d 列号码必须在 0–9 之间", i+3)
		}
		if (i == 6 || i == 7) && value > 127 {
			return nil, fmt.Errorf("第 %d 列超出 tinyint 范围", i+3)
		}
		if (i == 9 || i == 11 || i == 13) && value > 2147483647 {
			return nil, fmt.Errorf("第 %d 列注数超出 int 范围", i+3)
		}
		n[i] = value
	}
	return &models.Fucai3DRecord{Issue: fields[0], DrawDate: models.Date(date.Format("2006-01-02")),
		Num1: int(n[0]), Num2: int(n[1]), Num3: int(n[2]), TestNum1: int(n[3]), TestNum2: int(n[4]), TestNum3: int(n[5]), Machine: int(n[6]), BallSet: int(n[7]), TotalSales: n[8], DirectCount: int32(n[9]), DirectAmount: n[10], Group3Count: int32(n[11]), Group3Amount: n[12], Group6Count: int32(n[13]), Group6Amount: n[14]}, nil
}

type Fucai3DAnalysis struct {
	Numbers    []int  `json:"numbers"`
	Sum        int    `json:"sum"`
	Span       int    `json:"span"`
	OddCount   int    `json:"odd_count"`
	EvenCount  int    `json:"even_count"`
	BigCount   int    `json:"big_count"`
	SmallCount int    `json:"small_count"`
	Pattern    string `json:"pattern"`
}

func AnalyzeFucai3D(numbers []int) (*Fucai3DAnalysis, error) {
	if len(numbers) != 3 {
		return nil, fmt.Errorf("需要百、十、个位共 3 个数字")
	}
	r := &Fucai3DAnalysis{Numbers: append([]int(nil), numbers...)}
	min, max := 9, 0
	unique := map[int]bool{}
	for _, n := range numbers {
		if n < 0 || n > 9 {
			return nil, fmt.Errorf("每位数字必须在 0–9 之间")
		}
		r.Sum += n
		if n < min {
			min = n
		}
		if n > max {
			max = n
		}
		if n%2 == 0 {
			r.EvenCount++
		} else {
			r.OddCount++
		}
		if n >= 5 {
			r.BigCount++
		} else {
			r.SmallCount++
		}
		unique[n] = true
	}
	r.Span = max - min
	switch len(unique) {
	case 1:
		r.Pattern = "豹子"
	case 2:
		r.Pattern = "组三"
	case 3:
		r.Pattern = "组六"
	}
	return r, nil
}
