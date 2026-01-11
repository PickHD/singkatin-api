package helper

import (
	"os"
	"strconv"
)

func GetEnvString(e string) string {
	return os.Getenv(e)
}

func GetEnvInt(e string) int {
	eInt, _ := strconv.Atoi(os.Getenv(e))

	return eInt
}
