package lib

import (
	"gorm.io/gorm/clause"
)

// DBCompat provides database compatibility functions for MySQL and SQLite
type DBCompat struct{}

// NewDBCompat creates a new DBCompat instance
func NewDBCompat() DBCompat {
	return DBCompat{}
}

// Now returns the appropriate NOW() expression for the current database
// MySQL: NOW()
// SQLite: datetime('now', 'localtime')
func (DBCompat) Now() clause.Expr {
	if IsSQLite() {
		return clause.Expr{SQL: "datetime('now', 'localtime')"}
	}
	return clause.Expr{SQL: "NOW()"}
}

// Concat returns the appropriate string concatenation for the current database
// MySQL: CONCAT(a, b, c)
// SQLite: a || b || c
func (DBCompat) Concat(args ...string) string {
	if IsSQLite() {
		result := ""
		for i, arg := range args {
			if i > 0 {
				result += " || "
			}
			result += arg
		}
		return result
	}

	// MySQL CONCAT
	result := "CONCAT("
	for i, arg := range args {
		if i > 0 {
			result += ", "
		}
		result += arg
	}
	result += ")"
	return result
}

// TreePathLike returns the appropriate tree path LIKE condition
// This handles the pattern: CONCAT(',', tree_path, ',') LIKE '%,id,%'
// MySQL: CONCAT(',', tree_path, ',') LIKE '%,id,%'
// SQLite: (',' || tree_path || ',') LIKE '%,id,%'
func (DBCompat) TreePathLike(column string) string {
	if IsSQLite() {
		return "(',' || " + column + " || ',')"
	}
	return "CONCAT(',', " + column + ", ',')"
}
