package g

import (
	"fmt"
	"strings"

	"github.com/contextcloud/eventstore/es/filters"

	"gorm.io/gorm"
)

func whereClauseQuery(c filters.WhereClause) string {
	switch strings.ToLower(c.Op) {
	case `eq`:
		return fmt.Sprintf(`%s = ?`, c.Column)
	case `not.eq`:
		return fmt.Sprintf(`%s != ?`, c.Column)
	case `neq`:
		return fmt.Sprintf(`%s != ?`, c.Column)
	case `not.neq`:
		return fmt.Sprintf(`%s = ?`, c.Column)
	case `gt`:
		return fmt.Sprintf(`%s > ?`, c.Column)
	case `not.gt`:
		return fmt.Sprintf(`%s <= ?`, c.Column)
	case `gte`:
		return fmt.Sprintf(`%s >= ?`, c.Column)
	case `not.gte`:
		return fmt.Sprintf(`%s < ?`, c.Column)
	case `lt`:
		return fmt.Sprintf(`%s < ?`, c.Column)
	case `not.lt`:
		return fmt.Sprintf(`%s >= ?`, c.Column)
	case `lte`:
		return fmt.Sprintf(`%s <= ?`, c.Column)
	case `not.lte`:
		return fmt.Sprintf(`%s > ?`, c.Column)
	case `like`:
		return fmt.Sprintf(`%s ILIKE ?`, c.Column)
	case `not.like`:
		return fmt.Sprintf(`%s NOT ILIKE ?`, c.Column)
	case `is`:
		return fmt.Sprintf(`%s IS ?`, c.Column)
	case `not.is`:
		return fmt.Sprintf(`%s IS NOT ?`, c.Column)
	default:
		return ``
	}
}

func where(q *gorm.DB, filter filters.Where) *gorm.DB {
	switch w := filter.(type) {
	case []filters.Where:
		o := q.Session(&gorm.Session{NewDB: true})
		for _, inner := range w {
			o = where(o, inner)
		}
		return q.Where(o)
	case filters.WhereClause:
		return q.Where(whereClauseQuery(w), w.Args...)
	case filters.WhereOr:
		o := q.Session(&gorm.Session{NewDB: true})
		return q.Or(where(o, w.Where))
	default:
		return q
	}
}
