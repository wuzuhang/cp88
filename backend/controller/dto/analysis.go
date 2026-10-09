package dto

type AnalysisRequest struct {
	RedBalls []int `json:"red_balls" binding:"required,len=6"`
	BlueBall int   `json:"blue_ball" binding:"required,min=1,max=16"`
}

type AnalysisResponse struct {
	RedBalls   []int  `json:"red_balls"`
	BlueBall   int    `json:"blue_ball"`
	Sum        int    `json:"sum"`
	Span       int    `json:"span"`
	OddCount   int    `json:"odd_count"`
	EvenCount  int    `json:"even_count"`
	ZoneCounts [3]int `json:"zone_counts"`
}
