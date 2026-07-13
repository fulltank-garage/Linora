package models

type PageMetrics struct {
	Clicks      int64 `json:"clicks"`
	Engagements int64 `json:"engagements"`
	Impressions int64 `json:"impressions"`
	Reach       int64 `json:"reach"`
}

type DailyPageMetrics struct {
	Metrics    PageMetrics `json:"metrics"`
	RecordedOn string      `json:"recordedOn"`
}

type WeeklyReport struct {
	DaysWithData int         `json:"daysWithData"`
	EndDate      string      `json:"endDate"`
	Metrics      PageMetrics `json:"metrics"`
	StartDate    string      `json:"startDate"`
}

type FacebookPost struct {
	Comments  int64  `json:"comments"`
	CreatedAt string `json:"createdAt"`
	ID        string `json:"id"`
	Message   string `json:"message"`
	Reactions int64  `json:"reactions"`
	Shares    int64  `json:"shares"`
}

type PageSnapshot struct {
	Comments []ImportantComment `json:"-"`
	Metrics  PageMetrics        `json:"metrics"`
	Posts    []FacebookPost     `json:"posts"`
}

type ConnectedPageResponse struct {
	Page   FacebookPage   `json:"page"`
	Report AnalysisReport `json:"report"`
}
