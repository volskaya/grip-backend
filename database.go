package main

import (
	"log"
	"reflect"
	"sync"
	"time"

	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/lib/pq"
)

var (
	wg sync.WaitGroup
)

type Database struct {
	gorm *gorm.DB
}

type Migrations struct {
	User    *User
	Project *Project
	Raiting *Raiting
}

func migrate(db *gorm.DB) {
	log.Println("Running auto migrate…")

	migrations := Migrations{
		&User{},
		&Project{},
		&Raiting{},
	}

	v := reflect.ValueOf(migrations)
	wg.Add(v.NumField())

	for i := 0; i < v.NumField(); i++ {
		go func(i int) {
			defer wg.Done()
			db.AutoMigrate(v.Field(i).Interface())
		}(i)
	}

	wg.Wait()
	log.Println("Migrate complete")
}

func newDB(config *Config) *Database {
	var (
		orm       *gorm.DB
		connected bool
	)

	for !connected {
		log.Println("Establishing Database connection at postgres://" +
			config.PostgresHost + "/" + config.PostgresName + "…")

		db, err := gorm.Open(
			"postgres", "host="+config.PostgresHost+
				" user="+config.PostgresUser+
				" dbname="+config.PostgresName+
				" sslmode="+config.PostgresSSL+
				" password="+config.PostgresPassword)

		if err != nil {
			log.Println("Failed to connect to database at " +
				config.PostgresHost + "/" + config.PostgresName)

			time.Sleep(time.Second)
		} else {
			log.Println("Database connection established")
			migrate(db)

			connected = true
			orm = db
		}
	}

	return &Database{orm}
}

func (self *Database) close() {
	if self.gorm != nil {
		log.Println("Closing database connection")
		self.gorm.Close()
	}
}

func (self *Database) createUser(props *DiscordUser) *User {
	log.Printf("Creating user ID: %s, Name: %s", props.ID, props.Username)

	user := User{
		ID:            props.ID,
		Username:      props.Username,
		Avatar:        props.Avatar,
		Discriminator: props.Discriminator,
		Email:         props.Email,
	}

	self.gorm.Where(&User{ID: props.ID}).FirstOrCreate(&user)
	return &user
}

//------------------------------------------------------------------------------
func (self *Database) createProject(owner *User, props *Project) (*Project, error) {
	if self.gorm.NewRecord(props) {
		props.OwnerID = owner.ID
		props.RaitingIDs = pq.Int64Array{0, 0, 0, 0, 0}
		query := self.gorm.Create(props)

		log.Printf("Assigning Project ID: %d to User: %s\n", props.ID, owner.ID)
		owner.ProjectID = &props.ID
		self.gorm.First(owner).Update(owner)

		return props, query.Error
	}

	return props, nil
}

func (self *Database) deleteProject(id uint) error {
	req := self.gorm.Where("ID = ?", id).Delete(&Project{})
	return req.Error
}

//------------------------------------------------------------------------------
func (self *Database) createRaiting(
	ownerID *string,
	project *Project,
	raiting *Raiting,
) (*Raiting, error) {
	if self.gorm.NewRecord(raiting) {
		var user User

		if _, err := self.findID(&user, *ownerID); err != nil {
			return nil, fmt.Errorf("Tried to vote with a user ID, that does not exist: %s", *ownerID)
		}

		// If User has already Voted, update the previous vote instead
		var previousRaiting Raiting
		if !self.gorm.First(&previousRaiting, &Raiting{
			OwnerID:   *ownerID,
			ProjectID: project.ID,
		}).RecordNotFound() {
			log.Printf("Updating Users: %s, existing Raiting: %d, for Project: %d", *ownerID, previousRaiting.ID, project.ID)
			query := self.gorm.Model(&previousRaiting).Update(raiting)
			self.recalculateProjectRaiting(project)
			return &previousRaiting, query.Error
		}

		// Else proceed with a new Raiting
		log.Printf("Assigning Raiting ID: %d to Project: %d\n", raiting.ID, project.ID)
		raiting.OwnerID = *ownerID
		raiting.ProjectID = project.ID
		query := self.gorm.Create(raiting)

		self.recalculateProjectRaiting(project)
		return raiting, query.Error
	}

	return raiting, nil
}

func (self *Database) recalculateProjectRaiting(
	project *Project,
) bool {
	log.Printf("Recalculating Project raiting for %d", project.ID)

	var raitings []Raiting
	self.gorm.Model(project).Related(&raitings)

	var (
		n              = len(raitings)
		ids            = make(pq.Int64Array, n)
		design         int
		performance    int
		easeOfUse      int
		responsiveness int
		motion         int
	)

	if n == 0 {
		return true
	}

	for i, raiting := range raitings {
		ids[i] = int64(raiting.ID)
		design += raiting.Design
		performance += raiting.Performance
		easeOfUse += raiting.EaseOfUse
		responsiveness += raiting.Responsiveness
		motion += raiting.Motion
	}

	project.RaitingIDs = ids
	project.Raiting = pq.Int64Array{
		int64(design / n),
		int64(performance / n),
		int64(easeOfUse / n),
		int64(responsiveness / n),
		int64(motion / n),
	}

	query := self.gorm.Save(project)
	return query.Error == nil
}

// Universal versions
//------------------------------------------------------------------------------
func (self *Database) findID(ptr interface{}, id interface{}) (interface{}, error) {
	ref := reflect.ValueOf(ptr)
	req := self.gorm.First(ref.Interface(), "id = ?", id)
	return ref.Elem().Interface(), req.Error
}

func (self *Database) findWithID(ptr interface{}, ids []string) (interface{}, error) {
	ref := reflect.ValueOf(ptr)
	req := self.gorm.Where("ID in (?)", ids).Find(ref.Interface())
	return ref.Elem().Interface(), req.Error
}

func (self *Database) findAll(ptr interface{}) (interface{}, error) {
	ref := reflect.ValueOf(ptr)
	req := self.gorm.Find(ref.Interface())
	return ref.Elem().Interface(), req.Error
}
