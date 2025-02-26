package entities

type CommunityData struct {
	ID         int    `gorm:"primaryKey"`
	Day        string `gorm:"primaryKey"`
	Symbol     string
	Followers  string
	WatchCount string
}
