package guiio_util

import (
	"flag"
	"fmt"
	"time"
)

var (
	showVer bool
	Version string
)

func init() {
	flag.BoolVar(&showVer, "v", false, "show version")
}

func GenerateTRID(serverName string) string {
	return fmt.Sprintf("%s-%d", serverName, time.Now().UnixNano())
}

func ServerInfo(banner []byte) {
	fmt.Println(string(banner))
	flag.Parse()

	if showVer {
		fmt.Println("Version:", Version)
		return
	}
}
