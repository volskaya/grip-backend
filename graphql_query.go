package main

import (
	"context"
	"errors"
	"gopkg.in/validator.v2"
	"log"
	"reflect"
)

var (
	raitingPercentages = map[float64]int{
		0: 0,
		1: 25,
		2: 50,
		3: 75,
		4: 100,
	}
)

func validRaitingField(v interface{}, param string) error {
	ref := reflect.ValueOf(v)
	if ref.Kind() != reflect.Float64 {
		return errors.New("Unsupported type for inraitingrange validator")
	}

	switch ref.Float() {
	case 0:
	case 1:
	case 2:
	case 3:
	case 4:
		return nil
	default:
		return errors.New("This fields accepts either 0, 1, 2, 3 or 4")
	}

	return nil
}

type Query struct{}

// User
//------------------------------------------------------------------------------

func (_ *Query) User(ctx context.Context, args struct {
	ID *string
}) *User {
	if item, err := loadSomething(ctx, *args.ID, userLoaderKey); err == nil {
		if user, ok := item.(User); ok {
			return &user
		}
	}

	return nil
}

func (self *Query) Users(ctx context.Context) *Users {
	var users Users

	log.Println("Fetching all users")
	if _, err := ctx.Value("state").(*State).db.findAll(&users); err == nil {
		return &users
	}

	return nil
}

// Project queries
//------------------------------------------------------------------------------

func (_ *Query) Project(ctx context.Context, args struct {
	ID *string
}) *Project {
	if item, err := loadSomething(ctx, *args.ID, projectLoaderKey); err == nil {
		if project, ok := item.(Project); ok {
			return &project
		}
	}

	return nil
}

func (_ *Query) Projects(ctx context.Context) *Projects {
	var projects Projects

	log.Println("Fetching all projects")
	if _, err := ctx.Value("state").(*State).db.findAll(&projects); err == nil {
		return &projects
	}

	return nil
}

func (self *Query) NewProject(ctx context.Context, args struct {
	Link        string
	Github      string
	Description string
	Flags       string
	Picture     string
	Team        []string
	Theme       int32
}) *Project {
	// TODO: Validate fields
	// TODO: Convert base64 Picture to an actual picture and store its filename instead
	var (
		db         = ctx.Value("state").(*State).db
		authorized = ctx.Value("authorized").(bool)
	)

	if authorized {
		id := ctx.Value("user_id").(string)
		log.Printf("Creating a project for User %s\n", id)

		if item, err := loadSomething(ctx, id, userLoaderKey); err == nil {
			user := item.(User)

			// TODO: Convert the image and make the actual project
			// Creating project here
			if user.ProjectID != nil {
				// log.Printf("Tried to overlap a Project for User %s", user.ID)
				// return nil

				// FIXME: Use the above, when done with debug
				log.Printf("Deleting old Project %d for User %s\n", user.ProjectID, user.ID)
				db.deleteProject(*user.ProjectID)
			}

			if project, err := db.createProject(&user, &Project{
				Owner:       user,
				Link:        args.Link,
				Github:      args.Github,
				Description: args.Description,
				Flags:       args.Flags,
				Picture:     args.Picture,
				TeamUsers:   args.Team,
				Theme:       args.Theme,
			}); err == nil {
				log.Printf("Created project %d for User %s\n", project.ID, user.ID)
				return project
			}
		} else {
			log.Printf("Failed to fetch authorized User %s\n", id)
		}
	} else {
		log.Println("Tried to create a project for an unauthorized user")
	}

	return nil
}

func (self *Query) UpdateRaiting(ctx context.Context, args struct {
	ProjectID      string  `validate:"nonzero"`
	Design         float64 `validate:"min=0,max=100,validraiting"`
	Performance    float64 `validate:"min=0,max=100,validraiting"`
	EaseOfUse      float64 `validate:"min=0,max=100,validraiting"`
	Responsiveness float64 `validate:"min=0,max=100,validraiting"`
	Motion         float64 `validate:"min=0,max=100,validraiting"`
}) *Raiting {
	validator.SetValidationFunc("validraiting", validRaitingField)
	if err := validator.Validate(args); err != nil {
		log.Printf("Project %s updateRaiting validation failed, %s\n", args.ProjectID, err.Error())
		return nil
	}

	var (
		db         = ctx.Value("state").(*State).db
		authorized = ctx.Value("authorized").(bool)
	)

	if authorized {
		id := ctx.Value("user_id").(string)
		item, err := loadSomething(ctx, args.ProjectID, projectLoaderKey)

		if err != nil {
			return nil
		}

		project := item.(Project)
		log.Printf("Updating User's: %s vote on Project %d", id, project.ID)

		if raiting, err := db.createRaiting(&id, &project, &Raiting{
			Design:         raitingPercentages[args.Design],
			Performance:    raitingPercentages[args.Performance],
			EaseOfUse:      raitingPercentages[args.EaseOfUse],
			Responsiveness: raitingPercentages[args.Responsiveness],
			Motion:         raitingPercentages[args.Motion],
		}); err != nil {
			return raiting
		}
	}

	return nil
}
