package models

type ManualAnalysisInput struct {
	PageName          string `json:"pageName"`
	PostContent       string `json:"postContent"`
	Likes             int    `json:"likes"`
	Comments          int    `json:"comments"`
	Shares            int    `json:"shares"`
	ImportantComments string `json:"importantComments"`
	ExtraNotes        string `json:"extraNotes"`
}

type TopPost struct {
	PostID         string `json:"postId"`
	Excerpt        string `json:"excerpt"`
	Reason         string `json:"reason"`
	Recommendation string `json:"recommendation"`
}

type ImportantComment struct {
	CommentID      string `json:"commentId"`
	Message        string `json:"message"`
	Reason         string `json:"reason"`
	SuggestedReply string `json:"suggestedReply"`
}

type PostingDayInsight struct {
	AverageEngagement int64  `json:"averageEngagement"`
	Day               string `json:"day"`
	PostCount         int    `json:"postCount"`
}

type PostingTimeInsight struct {
	BasedOnPosts int                 `json:"basedOnPosts"`
	BestDay      string              `json:"bestDay"`
	BestTime     string              `json:"bestTime"`
	Days         []PostingDayInsight `json:"days"`
}

type AnalysisReport struct {
	ID                        string             `json:"id"`
	PageName                  string             `json:"pageName"`
	Summary                   string             `json:"summary"`
	SourceFingerprint         string             `json:"sourceFingerprint,omitempty"`
	HealthScore               int                `json:"healthScore"`
	TopPosts                  []TopPost          `json:"topPosts"`
	ImportantComments         []ImportantComment `json:"importantComments"`
	ContentRecommendations    []string           `json:"contentRecommendations"`
	AIContentRecommendation   string             `json:"aiContentRecommendation"`
	BestPostingTimes          []string           `json:"bestPostingTimes"`
	PostingTimeInsight        PostingTimeInsight `json:"postingTimeInsight"`
	PostingTimeRecommendation string             `json:"postingTimeRecommendation"`
	LineSummaryMessage        string             `json:"lineSummaryMessage"`
	Metrics                   PageMetrics        `json:"metrics"`
	CreatedAt                 string             `json:"createdAt"`
}
