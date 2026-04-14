package response

type Items struct {
	ID          string `json:"id"`
	WishListID  string `json:"wishlistId"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ProductURL  string `json:"productUrl"`
	Priority    int    `json:"priority"`
	Reserved    bool   `json:"reserved"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}
