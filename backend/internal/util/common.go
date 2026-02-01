package util

import (
	"flag"
	"fmt"
	"os"
	"time"
)

var (
	showVer bool
)

func init() {
	flag.BoolVar(&showVer, "v", false, "show version")
}

func GenerateTRID(serverName string, now time.Time) string {
	return fmt.Sprintf("%s-%d", serverName, now.UnixNano())
}

func ServerInfo(banner []byte, ver string) {
	fmt.Println(string(banner))
	flag.Parse()

	if showVer {
		fmt.Printf("Latest Version CheckOn GitCommit Hash \n✅ Version: %s\n", ver)
		os.Exit(0)
	}
}
