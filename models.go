package main

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	// graphql "github.com/graph-gophers/graphql-go"
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
	"golang.org/x/oauth2"
)

//------------------------------------------------------------------------------

type DiscordUser struct {
	Username      string  `json:"username"`
	ID            string  `json:"id"`
	Avatar        string  `json:"avatar"`
	Discriminator string  `json:"discriminator"`
	Email         *string `json:"email"`
}

type DiscordAuth struct {
	token *oauth2.Token
	user  *DiscordUser
}

type JwtClaims struct {
	Avatar        string `json:"avatar"`
	Discriminator string `json:"discriminator"`
	Username      string `json:"username"`
	jwt.StandardClaims
}

//------------------------------------------------------------------------------

type User struct {
	ID               string `gorm:"primary_key"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time `sql:"index"`
	Username         string
	Avatar           string
	Discriminator    string
	Email            *string
	IsAdmin          bool          `gorm:"default:false"`
	Voted            pq.Int64Array `gorm:"type:int[]"`
	Seen             pq.Int64Array `gorm:"type:int[]"`
	Visited          pq.Int64Array `gorm:"type:int[]"`
	Project          *Project
	ProjectID        *uint
	LastCommentCount int32
}

type Users []*User

type Project struct {
	gorm.Model
	Owner       User
	OwnerID     string
	Link        string
	Github      string
	Description string
	Flags       string
	Picture     string
	TeamUsers   pq.StringArray `gorm:"type:text[]"`
	Theme       int32
	Raiting     pq.Int64Array `gorm:"type:int[]"`
	RaitingIDs  pq.Int64Array `gorm:"type:int[]"`
	Raitings    []Raiting     `gorm:"foreignKey:ProjectID"`
}

type Projects []*Project

type Raiting struct {
	gorm.Model
	Owner          User
	OwnerID        string
	Project        Project
	ProjectID      uint
	Design         int
	Performance    int
	EaseOfUse      int
	Responsiveness int
	Motion         int
}

type Raitings []*Raiting
