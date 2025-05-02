package repository

type Client struct {
	ID       string `json:"id" db:"id"`
	Rate     int    `json:"rate_per_second" db:"rate"`
	Capacity int    `json:"capacity" db:"capacity"`
}
