package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"shuangseqiu/backend/models"
	"shuangseqiu/utils"
)

const maxDrawFileSize = 8 * 1024 * 1024

type recordWriter interface {
	Upsert(context.Context, *models.SSQRecord) error
}

type DrawSyncService struct {
	client    *http.Client
	sourceURL string
	store     recordWriter
}

func NewDrawSyncService(db *gorm.DB, sourceURL string) *DrawSyncService {
	return &DrawSyncService{
		client:    &http.Client{Timeout: 30 * time.Second},
		sourceURL: sourceURL,
		store:     &models.SSQRecordStore{DB: db},
	}
}

// SyncPreviousDay 只保存东八区前一日的开奖，不以文件最后一行代替目标日期。
func (s *DrawSyncService) SyncPreviousDay(ctx context.Context, now time.Time) (string, error) {
	local := now.In(utils.ChinaTimezone)
	target := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, utils.ChinaTimezone).AddDate(0, 0, -1)
	body, err := s.download(ctx)
	if err != nil {
		return "", err
	}
	record, err := parseDrawForDate(strings.NewReader(string(body)), target)
	if err != nil {
		return "", err
	}
	if err := s.store.Upsert(ctx, record); err != nil {
		return "", fmt.Errorf("保存期号 %s: %w", record.Issue, err)
	}
	return record.Issue, nil
}

func (s *DrawSyncService) download(ctx context.Context) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.sourceURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "shuangseqiu-backend/1.0")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("下载开奖文件: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("下载开奖文件: HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxDrawFileSize+1))
	if err != nil {
		return nil, fmt.Errorf("读取开奖文件: %w", err)
	}
	if len(body) > maxDrawFileSize {
		return nil, fmt.Errorf("开奖文件超过 %d 字节", maxDrawFileSize)
	}
	return body, nil
}

// FetchHistory 下载并校验全部历史数据，任何一行错误均阻止导入。
func (s *DrawSyncService) FetchHistory(ctx context.Context) ([]models.SSQRecord, error) {
	records, _, err := s.FetchHistorySnapshot(ctx)
	return records, err
}

// FetchHistorySnapshot 同时返回校验后的记录和原始文件，便于留存与离线重试。
func (s *DrawSyncService) FetchHistorySnapshot(ctx context.Context) ([]models.SSQRecord, []byte, error) {
	body, err := s.download(ctx)
	if err != nil {
		return nil, nil, err
	}
	records, err := ParseHistory(strings.NewReader(string(body)))
	if err != nil {
		return nil, nil, err
	}
	return records, body, nil
}

func ParseHistory(reader io.Reader) ([]models.SSQRecord, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	records := make([]models.SSQRecord, 0, 4000)
	seen := make(map[string]bool)
	line := 0
	for scanner.Scan() {
		line++
		fields := strings.Fields(strings.TrimPrefix(scanner.Text(), "\uFEFF"))
		if len(fields) == 0 {
			continue
		}
		record, err := parseDrawLine(fields)
		if err != nil {
			return nil, fmt.Errorf("历史数据第 %d 行: %w", line, err)
		}
		if seen[record.Issue] {
			return nil, fmt.Errorf("历史数据重复期号: %s", record.Issue)
		}
		seen[record.Issue] = true
		records = append(records, *record)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("历史数据为空")
	}
	return records, nil
}

func parseDrawForDate(reader io.Reader, target time.Time) (*models.SSQRecord, error) {
	date := target.In(utils.ChinaTimezone).Format("2006-01-02")
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	var result *models.SSQRecord
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		fields := strings.Fields(strings.TrimPrefix(scanner.Text(), "\uFEFF"))
		if len(fields) < 2 || fields[1] != date {
			continue
		}
		record, err := parseDrawLine(fields)
		if err != nil {
			return nil, fmt.Errorf("第 %d 行: %w", lineNumber, err)
		}
		if result != nil {
			return nil, fmt.Errorf("日期 %s 出现多条开奖记录", date)
		}
		result = record
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("扫描开奖文件: %w", err)
	}
	if result == nil {
		return nil, fmt.Errorf("开奖文件中尚无 %s 的开奖结果", date)
	}
	return result, nil
}

// 前 29 列对应数据库字段；31 列格式末尾的附加奖数据不写入本表。
func parseDrawLine(fields []string) (*models.SSQRecord, error) {
	if len(fields) != 29 && len(fields) != 31 {
		return nil, fmt.Errorf("开奖数据应为 29 或 31 列，实际 %d 列", len(fields))
	}
	if len(fields[0]) != 7 {
		return nil, fmt.Errorf("期号必须为 7 位数字")
	}
	for _, c := range fields[0] {
		if c < '0' || c > '9' {
			return nil, fmt.Errorf("期号必须为数字")
		}
	}
	drawDate, err := time.ParseInLocation("2006-01-02", fields[1], utils.ChinaTimezone)
	if err != nil {
		return nil, fmt.Errorf("开奖日期格式错误: %w", err)
	}
	numbers := make([]int64, 27)
	for i := range numbers {
		value, err := strconv.ParseInt(fields[i+2], 10, 64)
		if err != nil || value < 0 {
			return nil, fmt.Errorf("第 %d 列必须为非负整数", i+3)
		}
		numbers[i] = value
	}
	// 转换到 int 或 int32 之前检查数值范围，避免整数溢出。
	for i := 0; i < 13; i++ {
		max := int64(33)
		if i == 6 {
			max = 16
		}
		if numbers[i] < 1 || numbers[i] > max {
			return nil, fmt.Errorf("第 %d 列球号越界", i+3)
		}
	}
	for i := 15; i < 27; i += 2 {
		if numbers[i] > 2147483647 {
			return nil, fmt.Errorf("第 %d 列中奖注数超出 int 范围", i+3)
		}
	}
	reds := make([]int, 6)
	seq := make([]int, 6)
	for i := 0; i < 6; i++ {
		reds[i] = int(numbers[i])
		seq[i] = int(numbers[i+7])
	}
	if err := utils.ValidateBalls(reds, int(numbers[6])); err != nil {
		return nil, err
	}
	if err := utils.ValidateBalls(seq, int(numbers[6])); err != nil {
		return nil, fmt.Errorf("出球顺序: %w", err)
	}
	seen := make(map[int]bool, 6)
	for i, ball := range reds {
		if i > 0 && reds[i-1] >= ball {
			return nil, fmt.Errorf("排序红球必须严格递增")
		}
		seen[ball] = true
	}
	for _, ball := range seq {
		if !seen[ball] {
			return nil, fmt.Errorf("出球顺序与排序红球不一致")
		}
	}
	return &models.SSQRecord{
		Issue:        fields[0],
		DrawDate:     models.Date(drawDate.Format("2006-01-02")),
		Red1:         int(numbers[0]),
		Red2:         int(numbers[1]),
		Red3:         int(numbers[2]),
		Red4:         int(numbers[3]),
		Red5:         int(numbers[4]),
		Red6:         int(numbers[5]),
		Blue:         int(numbers[6]),
		SeqRed1:      int(numbers[7]),
		SeqRed2:      int(numbers[8]),
		SeqRed3:      int(numbers[9]),
		SeqRed4:      int(numbers[10]),
		SeqRed5:      int(numbers[11]),
		SeqRed6:      int(numbers[12]),
		TotalSales:   numbers[13],
		PoolAmount:   numbers[14],
		Prize1Count:  int32(numbers[15]),
		Prize1Amount: numbers[16],
		Prize2Count:  int32(numbers[17]),
		Prize2Amount: numbers[18],
		Prize3Count:  int32(numbers[19]),
		Prize3Amount: numbers[20],
		Prize4Count:  int32(numbers[21]),
		Prize4Amount: numbers[22],
		Prize5Count:  int32(numbers[23]),
		Prize5Amount: numbers[24],
		Prize6Count:  int32(numbers[25]),
		Prize6Amount: numbers[26],
	}, nil
}
