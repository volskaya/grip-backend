package main

import (
	"context"
	"fmt"
	"strconv"

	graphql "github.com/graph-gophers/graphql-go"
)

// User
//------------------------------------------------------------------------------

func (self User) Id() string {
	return self.ID
}

func (self User) NAME() string {
	return self.Username
}

func (self User) DISCRIMINATOR() string {
	return self.Discriminator
}

func (self User) AVATAR() string {
	return self.Avatar
}

func (self User) PROJECTID() *int32 {
	if self.ProjectID != nil {
		val := int32(*self.ProjectID)
		return &val
	}

	return nil
}

func (self User) PROJECT(ctx context.Context) *Project {
	if self.ProjectID != nil {
		if item, err := loadSomething(ctx, strconv.Itoa(int(*self.ProjectID)), projectLoaderKey); err == nil {
			project := item.(Project)
			return &project
		}
	}

	return nil
}

// Project
//------------------------------------------------------------------------------

func (self *Project) Id() graphql.ID {
	return graphql.ID(strconv.Itoa(int(self.ID)))
}

func (self *Project) String() string {
	return strconv.Itoa(int(self.ID))
}

func (self *Project) OWNER(ctx context.Context) *User {
	if self.OwnerID != "" {
		fmt.Printf("Fetching projects (%d) owner with ID %s\n", self.ID, self.OwnerID)
		if item, err := loadSomething(ctx, self.OwnerID, userLoaderKey); err == nil {
			owner := item.(User)
			return &owner
		}
	}

	return nil
}

func (self *Project) LINK() string {
	return self.Link
}

func (self *Project) GITHUB() string {
	return self.Github
}

func (self *Project) DESCRIPTION() string {
	return self.Description
}

func (self *Project) FLAGS() string {
	return self.Flags
}

func (self *Project) PICTURE() string {
	return self.Picture
}

func (self *Project) TEAM(ctx context.Context) []User {
	users := make([]User, len(self.TeamUsers))

	for i, id := range self.TeamUsers {
		if item, err := loadSomething(ctx, id, userLoaderKey); err == nil {
			users[i] = item.(User)
		}
	}

	return users
}

func (self *Project) THEME() int32 {
	return self.Theme
}

func (self *Project) RAITING() []int32 {
	raiting := make([]int32, len(self.Raiting))
	for i, val := range self.Raiting {
		raiting[i] = safeInt32(int(val))
	}

	return raiting
}

func (self *Project) RAITINGS(ctx context.Context) []Raiting {
	raitings := make([]Raiting, len(self.RaitingIDs))
	for i, id := range self.RaitingIDs {
		if item, err := loadSomething(ctx, strconv.Itoa(int(id)), raitingLoaderKey); err == nil {
			raitings[i] = item.(Raiting)
		}
	}

	return raitings
}

// Raiting
//------------------------------------------------------------------------------

func (self *Raiting) String() string {
	return strconv.Itoa(int(self.ID))
}

func (self Raiting) OWNER(ctx context.Context) *User {
	if self.OwnerID != "" {
		fmt.Printf("Fetching Raitings (%d) owner with ID %s\n", self.ID, self.OwnerID)
		if item, err := loadSomething(ctx, self.OwnerID, userLoaderKey); err == nil {
			owner := item.(User)
			return &owner
		}
	}

	return nil
}

func (self Raiting) PROJECT(ctx context.Context) *Project {
	fmt.Printf("Fetching raitings (%d) project with ID %s\n", self.ID, self.OwnerID)

	if item, err := loadSomething(ctx, strconv.Itoa(int(self.ProjectID)), projectLoaderKey); err == nil {
		project := item.(Project)
		return &project
	}

	return nil
}

func (self Raiting) DESIGN() int32 {
	return safeInt32(self.Design)
}
func (self Raiting) PERFORMANCE() int32 {
	return safeInt32(self.Performance)
}
func (self Raiting) EASEOFUSE() int32 {
	return safeInt32(self.EaseOfUse)
}
func (self Raiting) RESPONSIVENESS() int32 {
	return safeInt32(self.Responsiveness)
}
func (self Raiting) MOTION() int32 {
	return safeInt32(self.Motion)
}
