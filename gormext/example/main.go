package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/tenz-io/gokit/gormext"
	"github.com/tenz-io/gokit/logger"
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
	dsn := "root:mysql_123@tcp(localhost:3306)/gormext_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
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

	ctx := context.Background()
	err = Save(ctx, db, &User{
		Username: "admin",
		Password: "admin",
	})
	log.Printf("save user error: %+v\n", err)

	user, err := Find(ctx, db, "admin")
	log.Printf("find user error: %+v, user: %+v\n", err, user)

	user, err = Find(ctx, db, "sky")
	log.Printf("find user error: %+v, user: %+v\n", err, user)

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
	Username  string    `gorm:"column:username"`
	Password  string    `gorm:"column:password"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (d *User) TableName() string {
	return "user_tab"
}
