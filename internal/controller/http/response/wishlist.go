package response

type WishList struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	EventDate   string  `json:"eventDate"`
	PublicToken *string `json:"publicToken"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

type WishListWithItems struct {
	WishList
	Items []Items `json:"items"`
}
