package models

type FacebookPage struct {
	PageID   string `json:"pageId"`
	PageName string `json:"pageName"`
	Category string `json:"category"`
	IsActive bool   `json:"isActive"`
}
