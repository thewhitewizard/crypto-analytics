package entities

type Tweet struct {
	ConversationID string
	HTML           string
	ID             string `gorm:"primaryKey"`
	IsQuoted       bool
	IsPin          bool
	IsReply        bool
	IsRetweet      bool
	IsSelfThread   bool
	Likes          int
	Name           string
	Mentions       int
	PermanentURL   string
	Replies        int
	Retweets       int
	Text           string
	Timestamp      int64
	UserID         string
	Views          int
}
