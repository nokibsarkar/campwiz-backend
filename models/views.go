package models

import (
	"fmt"
	"slices"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type IView interface {
	schema.Tabler
	GetQuery(*gorm.DB) *gorm.DB
}

func MigrateViews(db *gorm.DB, views ...IView) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}
	if len(views) == 0 {
		return fmt.Errorf("no views provided for migration")
	}
	m := db.Migrator()
	tables, err := m.GetTables()
	if err != nil {
		return fmt.Errorf("failed to get tables: %w", err)
	}
	for _, view := range views {

		if !slices.Contains(tables, view.TableName()) {
			m.CreateView(view.TableName(), gorm.ViewOption{
				Query:   view.GetQuery(db),
				Replace: true,
			})
		} else {
			fmt.Printf("Replacing view %s\n", view.TableName())
			m.CreateView(view.TableName(), gorm.ViewOption{
				Query:   view.GetQuery(db),
				Replace: true,
			})
		}
	}
	return nil
}
