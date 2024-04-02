package main

import (
	"context"
	"fmt"
	syslog "log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/tenz-io/gokit/dbtracker"
	"github.com/tenz-io/gokit/logger"
)

func init() {
	logger.ConfigureWithOpts(
		logger.WithLoggerLevel(logger.DebugLevel),
		logger.WithConsoleEnabled(true),
		logger.WithFileEnabled(true),
	)

	logger.ConfigureTrafficWithOpts(
		logger.WithTrafficConsoleEnabled(true),
		logger.WithTrafficFileEnabled(true),
	)
}

func main() {
	dsn := "root:mysql_123@tcp(localhost:3306)/trackertest_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		syslog.Fatal("open database error: ", err)
		return
	}

	tracker := dbtracker.NewTrackerWithOpts(
		dbtracker.WithMetrics(false),
		dbtracker.WithTraffic(true),
		// test slow log, so set slow log floor to 1ms
		dbtracker.WithSlowLogFloor(1*time.Millisecond),
	)

	if err = setupTracking(db, tracker); err != nil {
		syslog.Fatal("setup tracking error: ", err)
		return
	}

	ctx := context.Background()
	err = Save(ctx, db, &User{
		Username: "admin",
		Password: "admin",
	})
	if err != nil {
		syslog.Println("save user error: ", err)
	}

	user, err := Find(ctx, db, "admin")
	if err != nil {
		syslog.Printf("find user error: %+v\n", err)
	}
	syslog.Printf("find user: %+v\n", user)

	time.Sleep(100 * time.Millisecond)

}

func setupTracking(db *gorm.DB, tracker dbtracker.Tracker) (err error) {
	err = db.Callback().Query().Before("*").Register("start_query", tracker.Start("db_query"))
	if err != nil {
		return fmt.Errorf("register start_query error: %w", err)
	}
	err = db.Callback().Query().After("*").Register("end_query", tracker.End())
	if err != nil {
		syslog.Fatal("register end_query error: ", err)
		return
	}

	err = db.Callback().Create().Before("*").Register("start_create", tracker.Start("db_create"))
	if err != nil {
		return fmt.Errorf("register start_create error: %w", err)
	}
	err = db.Callback().Create().After("*").Register("end_create", tracker.End())
	if err != nil {
		syslog.Fatal("register end_create error: ", err)
		return
	}

	err = db.Callback().Update().Before("*").Register("start_update", tracker.Start("db_update"))
	if err != nil {
		return fmt.Errorf("register start_update error: %w", err)
	}
	err = db.Callback().Update().After("*").Register("end_update", tracker.End())
	if err != nil {
		return fmt.Errorf("register end_update error: %w", err)
	}

	err = db.Callback().Delete().Before("*").Register("start_delete", tracker.Start("db_delete"))
	if err != nil {
		return fmt.Errorf("register start_delete error: %w", err)
	}
	err = db.Callback().Delete().After("*").Register("end_delete", tracker.End())
	if err != nil {
		return fmt.Errorf("register end_delete error: %w", err)
	}

	return nil
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
