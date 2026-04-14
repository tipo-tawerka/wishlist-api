package request

type SetWishlist struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description"`
	EventDate   string `json:"eventDate" validate:"required"`
}

type UpdateWishlist struct {
	Title       *string `json:"title" validate:"omitempty,min=1"`
	Description *string `json:"description"`
	EventDate   *string `json:"eventDate" validate:"omitempty,min=1"`
}
