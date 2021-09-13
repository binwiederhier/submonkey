package cmd

import (
	"errors"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
	"strings"
	"submonkey/submonkey"
	"submonkey/util"
)

// New creates a new CLI application
func New() *cli.App {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = os.TempDir()
	}
	cacheDir = filepath.Join(cacheDir, "submonkey")
	return &cli.App{
		Name:                   "submonkey",
		Usage:                  "Create videos from Reddit posts",
		UsageText:              "submonkey [OPTIONS..] OUTFILE.mp4",
		HideHelp:               true,
		HideVersion:            true,
		EnableBashCompletion:   true,
		UseShortOptionHandling: true,
		Reader:                 os.Stdin,
		Writer:                 os.Stdout,
		ErrWriter:              os.Stderr,
		Action:                 execCreate,
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "filter", Aliases: []string{"f"}, Value: "AnimalsBeingBros+AnimalsBeingDerps", Usage: "source subreddit(s) to be used for videos"},
			&cli.StringFlag{Name: "sort", Aliases: []string{"s"}, Value: "hot", Usage: "sort posts [hot, top, rising, new, controversial]"},
			&cli.StringFlag{Name: "time", Aliases: []string{"t"}, Value: "week", Usage: "time period to sort posts by [hour, day, week, month, year, all]"},
			&cli.IntFlag{Name: "limit", Aliases: []string{"l"}, Value: 5, Usage: "number of posts to include in the video"},
			&cli.BoolFlag{Name: "nsfw", Aliases: []string{"n"}, Usage: "include NSFW content"},
			&cli.StringFlag{Name: "size", Aliases: []string{"S"}, Value: "360p", Usage: "dimensions of the output video [360p, 720p, 1080p, WxH]"},
			&cli.StringFlag{Name: "cache-dir", Aliases: []string{"C"}, Value: cacheDir, Usage: "cache directory for video downloads"},
			&cli.StringFlag{Name: "cache-keep", Aliases: []string{"K"}, Value: "1d", Usage: "duration after which to delete cache entries"},
		},
	}
}

func execCreate(c *cli.Context) error {
	filter := c.String("filter")
	sort := c.String("sort")
	time := c.String("time")
	limit := c.Int("limit")
	size := c.String("size")
	nsfw := c.Bool("nsfw")
	cacheDir := c.String("cache-dir")
	cacheKeep := c.String("cache-keep")
	if c.NArg() < 1 {
		return errors.New("missing output file, see --help for usage details")
	} else if limit > 100 {
		return errors.New("limits must be <= 100")
	} else if !util.InStringList([]string{"hot", "top", "rising", "new", "controversial"}, sort) {
		return errors.New("sort must be any of: hot, top, rising, new, controversial")
	} else if !util.InStringList([]string{"hour", "day", "week", "month", "year", "all"}, time) {
		return errors.New("time must be any of: hour, day, week, month, year, all")
	}
	switch size {
	case "360p":
		size = "640:360"
	case "720p":
		size = "1280:720"
	case "1080p":
		size = "1920:1080"
	default:
		size = strings.ReplaceAll(size, "x", ":")
	}
	cacheKeepDuration, err := util.ParseDuration(cacheKeep)
	if err != nil {
		return err
	}
	filename := c.Args().Get(0)
	return submonkey.CreateVideo(&submonkey.CreateOptions{
		Filter:     filter,
		Sort:       sort,
		Time:       time,
		Limit:      limit,
		NSFW:       nsfw,
		OutputSize: size,
		OutputFile: filename,
		CacheDir:   cacheDir,
		CacheKeep:  cacheKeepDuration,
	})
}
