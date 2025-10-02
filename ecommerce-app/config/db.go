package config

import (
    "fmt"
    "time"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "github.com/sirupsen/logrus"
)

func InitDB(dsn string, logger *logrus.Logger) (*gorm.DB, error) {
    var db *gorm.DB
    var err error
    for i := 0; i < 5; i++ {
        db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
        if err == nil {
            sqlDB, _ := db.DB()
            if pingErr := sqlDB.Ping(); pingErr == nil {
                return db, nil
            }
            err = pingErr
        }
        logger.WithError(err).Warnf("Failed to connect to DB (attempt %d)", i+1)
        time.Sleep(2 * time.Second)
    }
    return nil, fmt.Errorf("could not connect to DB after retries: %w", err)
}

