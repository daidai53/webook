package dao

import "gorm.io/gorm"

func InitTables(db *gorm.DB) {
	db.AutoMigrate(&Payment{})
}
