// Copyright@daidai53 2024
package dao

import "gorm.io/gorm"

func InitTable(db *gorm.DB) error {
	return db.AutoMigrate(
		&Interactive{},
		&UserLikeBiz{},
		&UserCollectionBiz{},
	)
}
