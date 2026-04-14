package response

type PublicWishList struct {
	Title       string        `json:"title"`
	Description string        `json:"description"`
	EventDate   string        `json:"eventDate"`
	Items       []PublicItems `json:"items"`
}

type PublicItems struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ProductURL  string `json:"productUrl"`
	Priority    int    `json:"priority"`
	Reserved    bool   `json:"reserved"`
}
