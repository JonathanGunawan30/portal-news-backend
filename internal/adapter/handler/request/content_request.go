package request

type ContentRequest struct {
	Title       string `json:"title" validate:"required,min=3,max=200"`
	Excerpt     string `json:"excerpt" validate:"required,min=3,max=200"`
	Description string `json:"description" validate:"required,min=10"`
	Image       string `json:"image" validate:"required"`
	Tags        string `json:"tags" validate:"omitempty"`
	Status      string `json:"status" validate:"omitempty"`
	CategoryID  int64  `json:"category_id" validate:"required,gt=0"`
}
