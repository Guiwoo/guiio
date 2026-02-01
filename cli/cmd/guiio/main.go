package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"guiio/cli/internal/client"
)

var (
	apiBase = flag.String("api", "http://localhost:8080/api/v1", "API base URL")
	timeout = flag.Duration("timeout", 5*time.Second, "HTTP timeout")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "guiio CLI\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Commands:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  list\n  get <bucket>\n  create <bucket> [--region <region>]\n  delete <bucket>\n  upload <bucket> <file> [--name <object>] [--meta key=value]\n  download <bucket> <object> [--out <path>]\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	cmd := flag.Arg(0)
	rest := flag.Args()[1:]

	c := client.New(strings.TrimRight(*apiBase, "/"))
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	switch cmd {
	case "list":
		runList(ctx, c)
	case "get":
		runGet(ctx, c, rest)
	case "create":
		runCreate(ctx, c, rest)
	case "delete":
		runDelete(ctx, c, rest)
	case "upload":
		runUpload(ctx, c, rest)
	case "download":
		runDownload(ctx, c, rest)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		flag.Usage()
		os.Exit(1)
	}
}

func runList(ctx context.Context, c *client.Client) {
	resp, err := c.ListBuckets(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "list buckets failed: %v\n", err)
		os.Exit(1)
	}
	for _, b := range resp.Buckets {
		fmt.Printf("- %s (created %s)\n", b.Name, b.CreatedAt.Format(time.RFC3339))
	}
}

func runGet(ctx context.Context, c *client.Client, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "get requires bucket name")
		os.Exit(1)
	}
	resp, err := c.GetBucket(ctx, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "get bucket failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("%s\tregion=%s\tcreated=%s\n", resp.Name, resp.Region, resp.CreatedAt.Format(time.RFC3339))
}

func runCreate(ctx context.Context, c *client.Client, args []string) {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	region := fs.String("region", "", "Bucket region")
	_ = fs.Parse(args)
	remaining := fs.Args()
	if len(remaining) < 1 {
		fmt.Fprintln(os.Stderr, "create requires bucket name")
		os.Exit(1)
	}

	resp, err := c.CreateBucket(ctx, remaining[0], *region)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create bucket failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("created %s (region=%s)\n", resp.Name, resp.Region)
}

func runDelete(ctx context.Context, c *client.Client, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "delete requires bucket name")
		os.Exit(1)
	}

	resp, err := c.DeleteBucket(ctx, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "delete bucket failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("deleted %s\n", resp.Deleted)
}

func runUpload(ctx context.Context, c *client.Client, args []string) {
	fs := flag.NewFlagSet("upload", flag.ExitOnError)
	objectName := fs.String("name", "", "object name")
	meta := fs.String("meta", "", "metadata key=value comma separated")
	_ = fs.Parse(args)
	rest := fs.Args()
	if len(rest) < 2 {
		fmt.Fprintln(os.Stderr, "upload requires bucket and file path")
		os.Exit(1)
	}

	metaMap := map[string]string{}
	if *meta != "" {
		pairs := strings.Split(*meta, ",")
		for _, p := range pairs {
			if p == "" {
				continue
			}
			kv := strings.SplitN(p, "=", 2)
			if len(kv) == 2 {
				metaMap[kv[0]] = kv[1]
			}
		}
	}

	resp, err := c.UploadObject(ctx, rest[0], rest[1], *objectName, metaMap)
	if err != nil {
		fmt.Fprintf(os.Stderr, "upload failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("uploaded %s/%s size=%d etag=%s\n", resp.Bucket, resp.Object, resp.Size, resp.ETag)
}

func runDownload(ctx context.Context, c *client.Client, args []string) {
	fs := flag.NewFlagSet("download", flag.ExitOnError)
	outPath := fs.String("out", "", "output file path")
	_ = fs.Parse(args)
	rest := fs.Args()
	if len(rest) < 2 {
		fmt.Fprintln(os.Stderr, "download requires bucket and object")
		os.Exit(1)
	}
	bucket, object := rest[0], rest[1]
	res, err := c.DownloadObject(ctx, bucket, object)
	if err != nil {
		fmt.Fprintf(os.Stderr, "download failed: %v\n", err)
		os.Exit(1)
	}

	path := *outPath
	if path == "" {
		path = object
	}
	if err := os.WriteFile(path, res.Data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write file failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("downloaded %s (%d bytes) -> %s\n", object, len(res.Data), path)
}
