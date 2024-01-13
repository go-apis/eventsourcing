package filters

type Op string

const (
	OpEqual    Op = `eq`
	OpNotEqual Op = `not.eq`

	OpGreaterThan    Op = `gt`
	OpGreaterOrEqual Op = `gte`
	OpLessThan       Op = `lt`
	OpLessOrEqual    Op = `lte`
	OpLike           Op = `like`
	OpNotLike        Op = `not.like`
	OpIs             Op = `is`
	OpNotIs          Op = `not.is`
	OpIsNull         Op = `is.null`
	OpNotIsNull      Op = `not.is.null`
	OpIn             Op = `in`
	OpNotIn          Op = `not.in`
)
