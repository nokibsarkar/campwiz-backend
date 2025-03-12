package idgenerator

import (
	"nokib/campwiz/models"
	"strconv"

	"github.com/dimail777/snowflake-go"
)

// Create a new snowflake ID generator
var generator, _ = snowflake.InitByRandom()
var lastID int64 = 0

func GenerateID(prefix string) models.IDType {
	var n int64 = 0
	for n == 0 {
		n, _ = generator.GetNextId()
	}
	if n <= lastID {
		n = lastID + 1
	}
	lastID = n
	return models.IDType(prefix + strconv.FormatInt(n, 36))
}
func GenerateIDv2[T string](prefix string) T {
	var n int64
	for n == 0 {
		n, _ = generator.GetNextId()
	}
	return T(prefix + strconv.FormatInt(n, 36))
}
