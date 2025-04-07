package main

import (
	"fmt"
	"time"
)

func GenerateTRID(serverName string) string {
	return fmt.Sprintf("%s-%d", serverName, time.Now().UnixNano())
}
