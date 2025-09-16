package response

type ErrorResponseDefault struct {
	Meta
}

type Meta struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

type DefaultSuccessResponse struct {
	Meta       Meta                `json:"meta"`
	Data       interface{}         `json:"data,omitempty"`
	Pagination *PaginationResponse `json:"pagination,omitempty"`
}

type PaginationResponse struct {
	TotalRecord int64 `json:"total_record"`
	TotalPage   int64 `json:"total_page"`
	Page        int64 `json:"page"`
	PerPage     int64 `json:"per_page"`
}
