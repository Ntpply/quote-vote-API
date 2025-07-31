package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Quote struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Text      string             `bson:"text" json:"text"`
	Author    string             `bson:"author" json:"author"`         
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`   
	Votes     int                `bson:"votes" json:"votes"`
	VotedBy   []string           `bson:"votedBy" json:"votedBy"` // Added this field
	Category    string           `bson:"category" json:"category"`         

}
