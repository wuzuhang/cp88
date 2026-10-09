package service

import (
	"fmt"
	"sort"

	"shuangseqiu/backend/controller/dto"
	"shuangseqiu/utils"
)

type AnalysisService struct{}

func (s *AnalysisService) Analyze(req dto.AnalysisRequest) (*dto.AnalysisResponse, error) {
	if err := utils.ValidateBalls(req.RedBalls, req.BlueBall); err != nil {
		return nil, fmt.Errorf("号码校验失败: %w", err)
	}
	reds := append([]int(nil), req.RedBalls...)
	sort.Ints(reds)
	result := &dto.AnalysisResponse{RedBalls: reds, BlueBall: req.BlueBall, Span: reds[5] - reds[0]}
	for _, ball := range reds {
		result.Sum += ball
		if ball%2 == 1 {
			result.OddCount++
		} else {
			result.EvenCount++
		}
		result.ZoneCounts[(ball-1)/11]++
	}
	return result, nil
}
