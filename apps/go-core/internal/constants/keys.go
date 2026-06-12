package constants

import "time"

var (
	TTLShort  = 5 * time.Minute
	TTLMedium = 15 * time.Minute
	TTLLong   = 1 * time.Hour
)

const (
	PlansCacheKey = "plans:"
	ProductsKeys  = "products:"
	ServicesKeys  = "services:"
)
