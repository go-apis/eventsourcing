package filters

type Filter struct {
	Distinct []interface{}
	Where    Where
	Order    []Order
	Limit    *int
	Offset   *int
}

type Order struct {
	Column string
	Desc   bool
}

type Where interface{}

type WhereOr struct {
	Where
}

type WhereClause struct {
	Column string
	Op     Op
	Args   interface{}
}

func Limit(v int) *int {
	return &v
}
func Offset(v int) *int {
	return &v
}
