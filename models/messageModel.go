package models

type Message struct {
	MessageID  int    `form:"MessageID" gorm:"primaryKey;autoIncrement;column:MessageID"`
	SenderID   int    `json:"SenderID" gorm:"column:SenderID;not null"`
	ReceiverID int    `json:"ReceiverID" gorm:"column:ReceiverID;not null"`
	Text       string `json:"Text" gorm:"column:Text;type:text;not null"`
}
