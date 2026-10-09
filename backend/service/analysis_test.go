package service

import (
	"reflect"
	"testing"

	"shuangseqiu/backend/controller/dto"
)

func TestAnalyze(t *testing.T) {
	input := []int{33, 8, 19, 1, 25, 12}
	r, err := (&AnalysisService{}).Analyze(dto.AnalysisRequest{RedBalls: input, BlueBall: 6})
	if err != nil {
		t.Fatal(err)
	}
	if r.Sum != 98 || r.Span != 32 || r.OddCount != 4 || r.EvenCount != 2 || r.ZoneCounts != [3]int{2, 2, 2} {
		t.Fatalf("统计结果错误: %+v", r)
	}
	if !reflect.DeepEqual(r.RedBalls, []int{1, 8, 12, 19, 25, 33}) {
		t.Fatalf("红球未排序: %v", r.RedBalls)
	}
	if input[0] != 33 {
		t.Fatal("分析修改了输入")
	}
}

func TestAnalyzeRejectsInvalidBalls(t *testing.T) {
	cases := []dto.AnalysisRequest{
		{RedBalls: []int{1, 2, 3}, BlueBall: 1},
		{RedBalls: []int{1, 1, 3, 4, 5, 6}, BlueBall: 1},
		{RedBalls: []int{0, 2, 3, 4, 5, 6}, BlueBall: 1},
		{RedBalls: []int{1, 2, 3, 4, 5, 34}, BlueBall: 1},
		{RedBalls: []int{1, 2, 3, 4, 5, 6}, BlueBall: 0},
		{RedBalls: []int{1, 2, 3, 4, 5, 6}, BlueBall: 17},
	}
	for _, req := range cases {
		if _, err := (&AnalysisService{}).Analyze(req); err == nil {
			t.Errorf("非法号码被接受: %+v", req)
		}
	}
}
