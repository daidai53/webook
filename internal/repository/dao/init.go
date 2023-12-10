// Copyright@daidai53 2023
package dao

import "gorm.io/gorm"

func InitTable(db *gorm.DB) error {
	return db.AutoMigrate(&User{}, &Article{}, &PublishedArticle{})
}
