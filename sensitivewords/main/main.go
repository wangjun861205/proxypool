package main

import (
	"notbear/sensitivewords"
)

func main() {
	server := sensitivewords.NewServer()
	server.ListenAndServe()
}
