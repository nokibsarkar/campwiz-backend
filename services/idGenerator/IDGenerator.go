package idgenerator

import (
	"nokib/campwiz/models"
	"strconv"

	"github.com/dimail777/snowflake-go"
)

// Create a new snowflake ID generator
var generator, _ = snowflake.InitByRandom()

func GenerateID(prefix string) models.IDType {
	var n int64
	for n == 0 {
		n, _ = generator.GetNextId()
	}
	return models.IDType(prefix + strconv.FormatInt(n, 36))
}
