package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/tenz-io/gokit/gormext"
	"github.com/tenz-io/gokit/logger"
)

const (
	createUserTable string = `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE,
		password TEXT
	);`
)

func init() {
	logger.ConfigureWithOpts(
		logger.WithLoggerLevel(logger.DebugLevel),
		logger.WithSetAsDefaultLvl(true),
		logger.WithFileEnabled(true),
		logger.WithConsoleEnabled(true),
		logger.WithCallerEnabled(true),
		logger.WithCallerSkip(1),
	)

	logger.ConfigureTrafficWithOpts(
		logger.WithTrafficEnabled(true),
	)
}

func main() {
	db, err := gorm.Open(sqlite.Open("gorm_users.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("open database error: ", err)
		return
	}

	tracker := gormext.NewTrackerWithOpts(
		gormext.WithMetrics(false),
		gormext.WithTraffic(true),
		// test slow log, so set slow log floor to 1ms
		gormext.WithErrorLog(true),
		gormext.WithSlowLogFloor(1*time.Millisecond),
	)

	if err = tracker.Apply(db); err != nil {
		log.Fatal("setup tracking error: ", err)
		return
	}

	err = db.AutoMigrate(&User{})
	if err != nil {
		log.Fatal("auto migrate error: ", err)
		return
	}

	err = Save(context.Background(), db, &User{
		Username: "sky",
		Password: "sky123",
	})
	if err != nil {
		log.Printf("save user error: %+v\n", err)
		return
	}

	user, err := Find(context.Background(), db, "sky")
	if err != nil {
		log.Printf("find user error: %+v, user: %+v\n", err, user)
		return
	}
	log.Printf("find user: %+v\n", user)
	time.Sleep(100 * time.Millisecond)

}

func Find(ctx context.Context, db *gorm.DB, username string) (*User, error) {
	var (
		userModel User
		err       error
	)

	err = db.WithContext(ctx).
		Where("username = ?", username).
		First(&userModel).Error
	if err != nil {
		return nil, fmt.Errorf("find user error: %w", err)
	}

	return &userModel, nil
}

func Save(ctx context.Context, db *gorm.DB, user *User) error {
	err := db.WithContext(ctx).
		Create(user).Error
	if err != nil {
		return fmt.Errorf("save user error: %w", err)
	}

	return nil
}

type User struct {
	ID        int64     `gorm:"primaryKey"`
	Username  string    `gorm:"column:username;unique"`
	Password  string    `gorm:"column:password"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (d *User) TableName() string {
	return "user_tab"
}
