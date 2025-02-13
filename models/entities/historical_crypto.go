package entities

type Historical struct {
	ID        int     `json:"id,omitempty"`
	Slug      string  `json:"slug,omitempty" gorm:"primaryKey"`
	Day       string  `json:"day,omitempty" gorm:"primaryKey"`
	Symbol    string  `json:"symbol,omitempty"`
	Name      string  `json:"name,omitempty"`
	Price     float64 `json:"price"`
	Rank      int     `json:"cmcRank,omitempty"`
	Marketcap float64 `json:"marketCap"`
	Tags      string  `json:"tags,omitempty"`
}
