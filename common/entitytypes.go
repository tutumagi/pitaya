package common

import "fmt"

// EntityTableName 实体表名
func EntityTableName(entityTypName string) string {
	return fmt.Sprintf("tbl_%s", entityTypName)
}
