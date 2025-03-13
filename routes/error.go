package routes

import "log"

func HandleError(endpoint string) {
	// Handle error here
	if r := recover(); r != nil {
		// log the error
		log.Println("Error at endpoint", endpoint, ":", r)
	}
}
