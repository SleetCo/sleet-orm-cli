// Package puller introspects a live MySQL/MariaDB database and converts
// the information_schema into Sleet TableDef structs.
package puller

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/SleetCo/sleet-orm-cli/internal/loader"
)

// Config holds the database connection parameters.
type Config struct {
	Host   string
	Port   int
	User   string
	Pass   string
	DBName string
}

// Pull connects to the database and returns all user table definitions.
func Pull(cfg Config) ([]*loader.TableDef, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		cfg.User, cfg.Pass, cfg.Host, cfg.Port, cfg.DBName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open connection: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping %s:%d: %w", cfg.Host, cfg.Port, err)
	}

	return introspect(db, cfg.DBName)
}

func introspect(db *sql.DB, dbName string) ([]*loader.TableDef, error) {
	rows, err := db.Query(`
		SELECT TABLE_NAME
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = ? AND TABLE_TYPE = 'BASE TABLE'
		ORDER BY TABLE_NAME
	`, dbName)
	if err != nil {
		return nil, fmt.Errorf("list tables: %w", err)
	}
	defer rows.Close()

	var tables []*loader.TableDef
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		cols, err := introspectColumns(db, dbName, name)
		if err != nil {
			return nil, fmt.Errorf("columns of %s: %w", name, err)
		}
		tables = append(tables, &loader.TableDef{Name: name, Columns: cols})
	}

	return tables, rows.Err()
}

func introspectColumns(db *sql.DB, dbName, tableName string) ([]*loader.ColumnDef, error) {
	rows, err := db.Query(`
		SELECT
			COLUMN_NAME,
			DATA_TYPE,
			COLUMN_TYPE,
			CHARACTER_MAXIMUM_LENGTH,
			NUMERIC_PRECISION,
			NUMERIC_SCALE,
			IS_NULLABLE,
			COLUMN_KEY,
			EXTRA,
			COLUMN_DEFAULT
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`, dbName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []*loader.ColumnDef
	for rows.Next() {
		var (
			name       string
			dataType   string
			colType    string
			charLen    sql.NullInt64
			numPrec    sql.NullInt64
			numScale   sql.NullInt64
			isNullable string
			colKey     string
			extra      string
			colDefault sql.NullString
		)

		if err := rows.Scan(
			&name, &dataType, &colType,
			&charLen, &numPrec, &numScale,
			&isNullable, &colKey, &extra, &colDefault,
		); err != nil {
			return nil, err
		}

		col := &loader.ColumnDef{
			Name:          name,
			Type:          normalizeType(dataType, colType),
			NotNull:       isNullable == "NO",
			PrimaryKey:    colKey == "PRI",
			Unique:        colKey == "UNI",
			AutoIncrement: strings.Contains(extra, "auto_increment"),
		}

		if charLen.Valid {
			col.Length = int(charLen.Int64)
		}
		if numPrec.Valid {
			col.Precision = int(numPrec.Int64)
		}
		if numScale.Valid {
			col.Scale = int(numScale.Int64)
		}

		if colDefault.Valid {
			upper := strings.ToUpper(colDefault.String)
			if strings.Contains(upper, "CURRENT_TIMESTAMP") || upper == "NOW()" {
				col.DefaultNow = true
				col.HasDefault = true
			} else {
				s := colDefault.String
				col.Default = &s
				col.HasDefault = true
			}
		}

		cols = append(cols, col)
	}

	return cols, rows.Err()
}

// normalizeType converts MySQL's DATA_TYPE / COLUMN_TYPE to our canonical type name.
func normalizeType(dataType, colType string) string {
	dt := strings.ToUpper(dataType)

	// tinyint(1) is MySQL's boolean idiom
	if strings.EqualFold(colType, "tinyint(1)") {
		return "BOOLEAN"
	}

	switch dt {
	case "INT", "INTEGER":
		return "INT"
	case "TINYINT":
		return "TINYINT"
	case "SMALLINT":
		return "SMALLINT"
	case "BIGINT":
		return "BIGINT"
	case "FLOAT":
		return "FLOAT"
	case "DOUBLE", "DOUBLE PRECISION":
		return "DOUBLE"
	case "DECIMAL", "NUMERIC":
		return "DECIMAL"
	case "CHAR":
		return "CHAR"
	case "VARCHAR":
		return "VARCHAR"
	case "TEXT":
		return "TEXT"
	case "MEDIUMTEXT":
		return "MEDIUMTEXT"
	case "LONGTEXT":
		return "LONGTEXT"
	case "JSON":
		return "JSON"
	case "TIMESTAMP":
		return "TIMESTAMP"
	case "DATETIME":
		return "DATETIME"
	case "DATE":
		return "DATE"
	default:
		return dt
	}
}
