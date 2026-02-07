package lib

import (
	"gorm.io/gorm/clause"
)

// DBCompat provides database compatibility functions for MySQL, SQLite and PostgreSQL
type DBCompat struct{}

// NewDBCompat creates a new DBCompat instance
func NewDBCompat() DBCompat {
	return DBCompat{}
}

// Now returns the appropriate NOW() expression for the current database
// MySQL: NOW()
// PostgreSQL: NOW()
// SQLite: datetime('now', 'localtime')
func (DBCompat) Now() clause.Expr {
	if IsSQLite() {
		return clause.Expr{SQL: "datetime('now', 'localtime')"}
	}
	// MySQL and PostgreSQL both support NOW()
	return clause.Expr{SQL: "NOW()"}
}

// Concat returns the appropriate string concatenation for the current database
// MySQL: CONCAT(a, b, c)
// PostgreSQL: a || b || c
// SQLite: a || b || c
func (DBCompat) Concat(args ...string) string {
	if IsSQLite() || IsPostgreSQL() {
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
// PostgreSQL: (',' || tree_path || ',') LIKE '%,id,%'
// SQLite: (',' || tree_path || ',') LIKE '%,id,%'
func (DBCompat) TreePathLike(column string) string {
	if IsSQLite() || IsPostgreSQL() {
		return "(',' || " + column + " || ',')"
	}
	return "CONCAT(',', " + column + ", ',')"
}
