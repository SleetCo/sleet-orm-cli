// Package loader executes a Sleet schema.lua file inside a gopher-lua VM
// and extracts all table/column definitions without needing a real database.
//
// The trick: before running the user's schema file, we register a mock "sleet"
// module. Every sl.table() / sl.int() / sl.varchar():notNull() call is
// intercepted and stored in Go structs. No regex parsing required.
package loader

import (
	"fmt"
	"sort"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

// ColumnDef is the Go representation of a column defined in schema.lua.
type ColumnDef struct {
	Name          string
	Order         int // for order sort
	Type          string // SQL type string, e.g. "INT", "VARCHAR", "BOOLEAN"
	Length        int    // for VARCHAR / CHAR
	Precision     int    // for DECIMAL
	Scale         int    // for DECIMAL
	NotNull       bool
	PrimaryKey    bool
	Unique        bool
	AutoIncrement bool
	HasDefault    bool
	Default       *string // nil means no default
	DefaultNow    bool
	References    *string // "tableName.colName" or nil
	Comment       string  // column description from .comment("..."), empty if not set
}

// TableDef is the Go representation of a sl.table() call.
type TableDef struct {
	Name    string
	Columns []*ColumnDef
}

// LoadSchema runs the given schema.lua file in an isolated Lua VM,
// intercepts all sl.table() calls, and returns the collected definitions.
func LoadSchema(path string) ([]*TableDef, error) {
	L := lua.NewState(lua.Options{SkipOpenLibs: false})
	defer L.Close()

	var tables []*TableDef

	// Preload our mock "sleet" module so that require('sleet') in the
	// schema file returns our interceptor table instead of the real library.
	L.PreloadModule("sleet", func(L *lua.LState) int {
		mod := buildMockModule(L, &tables)
		L.Push(mod)
		return 1
	})

	// Also set _G.Sleet so schemas that do `local sl = Sleet` work too.
	mock := buildMockModule(L, &tables)
	L.SetGlobal("Sleet", mock)

	if err := L.DoFile(path); err != nil {
		return nil, fmt.Errorf("executing %s: %w", path, err)
	}

	return tables, nil
}

// buildMockModule creates the Lua table that mimics the real Sleet module.
func buildMockModule(L *lua.LState, tables *[]*TableDef) *lua.LTable {
	mod := L.NewTable()

	// ── Column type constructors ──────────────────────────────────────────
	type colSpec struct {
		sqlType       string
		autoIncrement bool
	}

	specs := map[string]colSpec{
		"serial":     {"INT", true},
		"bigserial":  {"BIGINT", true},
		"int":        {"INT", false},
		"bigint":     {"BIGINT", false},
		"smallint":   {"SMALLINT", false},
		"tinyint":    {"TINYINT", false},
		"float":      {"FLOAT", false},
		"double":     {"DOUBLE", false},
		"text":       {"TEXT", false},
		"mediumtext": {"MEDIUMTEXT", false},
		"longtext":   {"LONGTEXT", false},
		"boolean":    {"BOOLEAN", false},
		"timestamp":  {"TIMESTAMP", false},
		"datetime":   {"DATETIME", false},
		"date":       {"DATE", false},
		"json":       {"JSON", false},
	}

	sequence := 0
    getNextOrder := func() int {
        sequence++
        return sequence
    }

	for typeName, spec := range specs {
		typeName, spec := typeName, spec
		L.SetField(mod, typeName, L.NewFunction(func(L *lua.LState) int {
			// These are called as sl.int(), sl.text() etc. — no self arg (dot syntax)
			length := L.OptInt(1, 0)
			col := &ColumnDef{
				Type:          spec.sqlType,
				Length:        length,
				AutoIncrement: spec.autoIncrement,
				Order:         getNextOrder(),
			}
			L.Push(buildColTable(L, col))
			return 1
		}))
	}

	// varchar(len) and char(len) accept an explicit length
	for _, typeName := range []string{"varchar", "char"} {
		typeName := typeName
		sqlType := strings.ToUpper(typeName)
		defaultLen := 255
		if typeName == "char" {
			defaultLen = 1
		}
		L.SetField(mod, typeName, L.NewFunction(func(L *lua.LState) int {
			length := L.OptInt(1, defaultLen)
			col := &ColumnDef{Type: sqlType, Length: length, Order: getNextOrder() }
			L.Push(buildColTable(L, col))
			return 1
		}))
	}

	// decimal(precision, scale)
	L.SetField(mod, "decimal", L.NewFunction(func(L *lua.LState) int {
		precision := L.OptInt(1, 10)
		scale := L.OptInt(2, 2)
		col := &ColumnDef{Type: "DECIMAL", Precision: precision, Scale: scale, Order: getNextOrder() }
		L.Push(buildColTable(L, col))
		return 1
	}))

	// ── sl.table(name, columns) ───────────────────────────────────────────
	L.SetField(mod, "table", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)
		colsLua := L.CheckTable(2)

		tbl := &TableDef{Name: name}

		colsLua.ForEach(func(key lua.LValue, val lua.LValue) {
			colName, ok := key.(lua.LString)
			if !ok {
				return
			}
			// Skip internal/private keys
			if strings.HasPrefix(string(colName), "_") {
				return
			}
			colLuaTable, ok := val.(*lua.LTable)
			if !ok {
				return
			}
			ud, ok := L.GetField(colLuaTable, "__sleetCol").(*lua.LUserData)
			if !ok {
				return
			}
			col, ok := ud.Value.(*ColumnDef)
			if !ok {
				return
			}
			col.Name = string(colName)
			tbl.Columns = append(tbl.Columns, col)
		})

		sort.Slice(tbl.Columns, func(i, j int) bool {
			return tbl.Columns[i].Order < tbl.Columns[j].Order
		})

		*tables = append(*tables, tbl)

		// Return a proxy table so that columns can be referenced
		// for :references() — e.g. users.identifier
		proxy := L.NewTable()
		L.SetField(proxy, "_tableName", lua.LString(name))
		for _, col := range tbl.Columns {
			col := col
			colProxy := L.NewTable()
			L.SetField(colProxy, "_name", lua.LString(col.Name))
			L.SetField(colProxy, "_tableName", lua.LString(name))
			L.SetField(proxy, col.Name, colProxy)
		}

		L.Push(proxy)
		return 1
	}))

	// ── sl.sql() stub (no-op for schema parsing) ──────────────────────────
	L.SetField(mod, "sql", L.NewFunction(func(L *lua.LState) int {
		frag := L.OptString(1, "")
		t := L.NewTable()
		L.SetField(t, "_fragment", lua.LString(frag))
		L.Push(t)
		return 1
	}))

	// ── Stub all operator functions (eq, ne, gt, …) ───────────────────────
	for _, opName := range []string{
		"eq", "ne", "gt", "gte", "lt", "lte",
		"like", "ilike", "notLike",
		"isNull", "isNotNull",
		"inArray", "notInArray",
		"between", "and_", "or_", "not_",
	} {
		L.SetField(mod, opName, L.NewFunction(func(L *lua.LState) int {
			L.Push(L.NewTable()) // return empty table — won't be used during schema loading
			return 1
		}))
	}

	return mod
}

// buildColTable wraps a *ColumnDef in a Lua table and attaches all chainable methods.
func buildColTable(L *lua.LState, col *ColumnDef) *lua.LTable {
	t := L.NewTable()

	// Embed the Go pointer as userdata so sl.table() can retrieve it
	ud := L.NewUserData()
	ud.Value = col
	L.SetField(t, "__sleetCol", ud)

	// Helper: chain method that sets a flag and returns self
	chain := func(name string, fn func()) {
		L.SetField(t, name, L.NewFunction(func(L *lua.LState) int {
			// L.Get(1) is self (colon-call), we use closure vars
			fn()
			L.Push(t)
			return 1
		}))
	}

	chain("primaryKey", func() { col.PrimaryKey = true; col.NotNull = true })
	chain("notNull", func() { col.NotNull = true })
	chain("unique", func() { col.Unique = true })
	chain("autoIncrement", func() { col.AutoIncrement = true })
	chain("defaultNow", func() { col.DefaultNow = true; col.HasDefault = true })

	// .default(val) — dot-notation: first arg IS the value (no self)
	L.SetField(t, "default", L.NewFunction(func(L *lua.LState) int {
		val := L.Get(1)
		var s string
		switch v := val.(type) {
		case lua.LBool:
			if bool(v) {
				s = "1"
			} else {
				s = "0"
			}
		case lua.LNumber:
			s = lua.LVAsString(v)
		case lua.LString:
			s = string(v)
		default:
			s = val.String()
		}
		col.Default = &s
		col.HasDefault = true
		L.Push(t)
		return 1
	}))

	// .references(otherCol) — dot-notation: first arg IS the column proxy
	L.SetField(t, "references", L.NewFunction(func(L *lua.LState) int {
		ref, ok := L.Get(1).(*lua.LTable)
		if ok {
			tblName := lua.LVAsString(L.GetField(ref, "_tableName"))
			colName := lua.LVAsString(L.GetField(ref, "_name"))
			if tblName != "" && colName != "" {
				refStr := tblName + "." + colName
				col.References = &refStr
			}
		}
		L.Push(t)
		return 1
	}))

	// .comment(text) — stores the description string for type generation and SQL
	L.SetField(t, "comment", L.NewFunction(func(L *lua.LState) int {
		col.Comment = L.OptString(1, "")
		L.Push(t)
		return 1
	}))

	return t
}
