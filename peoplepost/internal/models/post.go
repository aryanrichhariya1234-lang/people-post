package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Location struct {
	Lat float64 `bson:"lat,omitempty" json:"lat"`
	Lng float64 `bson:"lng,omitempty" json:"lng"`
}

type Like struct {
	User primitive.ObjectID `bson:"user" json:"user"`
}

type Post struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	User        primitive.ObjectID `bson:"user" json:"user"`
	Category    string             `bson:"category" json:"category"`
	Address     string             `bson:"Address" json:"Address"`
	Description string             `bson:"description" json:"description"`
	Images      []string           `bson:"images" json:"images"`
	Location    Location           `bson:"location" json:"location"`
	Status      string             `bson:"status,omitempty" json:"status"`
	Likes       []Like             `bson:"likes" json:"likes"`
	CreatedAt   time.Time          `bson:"createdAt,omitempty" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt,omitempty" json:"updatedAt"`
}