package domain

type PaginationQuery struct {
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
	Sort  string `json:"sort,omitempty"`
	Order string `json:"order,omitempty"`
}

type PaginatedResult struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalRows  int64       `json:"total_rows"`
	TotalPages int         `json:"total_pages"`
}