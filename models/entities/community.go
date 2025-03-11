package entities

type CommunityData struct {
	Cid        int    `gorm:"primaryKey"`
	Day        string `gorm:"primaryKey"`
	Symbol     string `gorm:"primaryKey"`
	Followers  string
	WatchCount string
}
