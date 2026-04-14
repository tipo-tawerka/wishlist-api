package request

type SetItems struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description"`
	ProductUrl  string `json:"productUrl" validate:"omitempty,url"`
	Priority    int    `json:"priority" validate:"required,min=1,max=5"`
}

type UpdateItems struct {
	Title       *string `json:"title" validate:"omitempty,min=1"`
	Description *string `json:"description"`
	ProductUrl  *string `json:"productUrl" validate:"omitempty,url"`
	Priority    *int    `json:"priority" validate:"omitempty,min=1,max=5"`
}
