package dto

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/guregu/null"
)

type Filters struct {
	Column     null.String `json:"Column" swaggertype:"string"`
	Descending bool        `json:"Descending"`
	Count      null.Int    `json:"Count" swaggertype:"integer"`
	Offset     null.Int    `json:"Offset" swaggertype:"integer"`
}

func (f Filters) Validate(entity interface{}) error {
	if !f.Column.Valid {
		return nil
	}
	t := reflect.TypeOf(entity)
	for i := 0; i < t.NumField(); i++ {
		if strings.EqualFold(f.Column.String, t.Field(i).Name) {
			return nil
		}
	}
	return fmt.Errorf("Column '%s' is not a valid filter for %s entity", f.Column.String, t.Name())
}

func (f Filters) GetQuery() string {
	var sb = &strings.Builder{}
	if f.Column.Valid {
		fmt.Fprintf(sb, " ORDER BY %s", f.Column.String)
		if f.Descending {
			fmt.Fprintf(sb, " DESC")
		} else {
			fmt.Fprintf(sb, " ASC")
		}
	}
	if f.Count.Valid {
		fmt.Fprintf(sb, " LIMIT %d", f.Count.Int64)
	}
	if f.Offset.Valid && f.Offset.Int64 > 0 {
		fmt.Fprintf(sb, " OFFSET %d", f.Offset.Int64)
	}
	return sb.String()
}
