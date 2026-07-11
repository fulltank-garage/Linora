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
	Reason         string `json:"reason"`
	Recommendation string `json:"recommendation"`
}

type ImportantComment struct {
	CommentID      string `json:"commentId"`
	Message        string `json:"message"`
	Reason         string `json:"reason"`
	SuggestedReply string `json:"suggestedReply"`
}

type AnalysisReport struct {
	ID                     string             `json:"id"`
	PageName               string             `json:"pageName"`
	Summary                string             `json:"summary"`
	HealthScore            int                `json:"healthScore"`
	TopPosts               []TopPost          `json:"topPosts"`
	ImportantComments      []ImportantComment `json:"importantComments"`
	ContentRecommendations []string           `json:"contentRecommendations"`
	BestPostingTimes       []string           `json:"bestPostingTimes"`
	LineSummaryMessage     string             `json:"lineSummaryMessage"`
	Metrics                PageMetrics        `json:"metrics"`
	CreatedAt              string             `json:"createdAt"`
}
