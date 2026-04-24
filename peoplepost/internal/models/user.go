package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name               string             `bson:"name" json:"name"`
	Email              string             `bson:"email" json:"email"`
	Password           string             `bson:"password,omitempty" json:"-"`
	PasswordChangedAt  *time.Time         `bson:"passwordChangedAt,omitempty" json:"passwordChangedAt,omitempty"`
	PasswordResetToken string             `bson:"passwordResetToken,omitempty" json:"passwordResetToken,omitempty"`
	PasswordResetTime  *time.Time         `bson:"passwordResetTime,omitempty" json:"passwordResetTime,omitempty"`
	Role               string             `bson:"role" json:"role"`
	GovernmentID       string             `bson:"governmentId,omitempty" json:"governmentId,omitempty"`
	Active             bool               `bson:"active,omitempty" json:"active,omitempty"`
	Photo              string             `bson:"photo,omitempty" json:"photo,omitempty"`
	PublicID           string             `bson:"public_id,omitempty" json:"public_id,omitempty"`
}