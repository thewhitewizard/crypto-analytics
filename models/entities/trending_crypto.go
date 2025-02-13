package entities

type TrendingCrypto struct {
	ID     int    `json:"id,omitempty"`
	Slug   string `json:"slug,omitempty" gorm:"primaryKey"`
	Day    string `json:"day,omitempty" gorm:"primaryKey"`
	Symbol string `json:"symbol,omitempty"`
	Name   string `json:"name,omitempty"`
}
