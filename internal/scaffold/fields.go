package scaffold

import (
	"fmt"
	"strings"
)

// Field is one attribute of a generated domain, derived from a "--fields"
// entry such as "favorite_food:string".
type Field struct {
	Exported   string // FavoriteFood
	Unexported string // favoriteFood
	JSONTag    string // favorite_food
	SQLColumn  string // favorite_food
	GoType     string // string
	SQLType    string // text
	ScanZero   string // zero value literal used when declaring a scan target
}

var goTypeToSQL = map[string]string{
	"string":    "text",
	"int":       "integer",
	"bool":      "boolean",
	"float":     "double precision",
	"time":      "timestamptz",
	"time.Time": "timestamptz",
}

var fieldTypeAliases = map[string]string{
	"string": "string",
	"str":    "string",
	"text":   "string",
	"int":    "int",
	"integer": "int",
	"bool":    "bool",
	"boolean": "bool",
	"float":   "float64",
	"float64": "float64",
	"time":    "time.Time",
	"time.Time": "time.Time",
	"timestamp": "time.Time",
}

var goZeroValue = map[string]string{
	"string":    `""`,
	"int":       "0",
	"bool":      "false",
	"float64":   "0",
	"time.Time": "time.Time{}",
}

// ParseFields parses a Rails-generator-style field spec, e.g.
// "name:string,color:string,legs:int", into a list of Fields ready to drive
// the domain templates. An empty spec yields no fields (a bare skeleton).
func ParseFields(spec string) ([]Field, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil, nil
	}

	var fields []Field
	for _, raw := range strings.Split(spec, ",") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}

		parts := strings.SplitN(raw, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid field %q: expected name:type", raw)
		}

		name := strings.TrimSpace(parts[0])
		rawType := strings.ToLower(strings.TrimSpace(parts[1]))
		if name == "" || rawType == "" {
			return nil, fmt.Errorf("invalid field %q: expected name:type", raw)
		}

		goType, ok := fieldTypeAliases[rawType]
		if !ok {
			return nil, fmt.Errorf("field %q: unsupported type %q (supported: string, int, bool, float, time)", name, rawType)
		}

		snake := SnakeCase(name)
		fields = append(fields, Field{
			Exported:   PascalCase(name),
			Unexported: CamelCase(name),
			JSONTag:    snake,
			SQLColumn:  snake,
			GoType:     goType,
			SQLType:    goTypeToSQL[goType],
			ScanZero:   goZeroValue[goType],
		})
	}

	return fields, nil
}
