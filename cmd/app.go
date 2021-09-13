package cmd

import (
	"errors"
	"github.com/urfave/cli/v2"
	"os"
	"strings"
	"submonkey/submonkey"
	"submonkey/util"
)

// New creates a new CLI application
func New() *cli.App {
	return &cli.App{
		Name:                   "submonkey",
		Usage:                  "Create video from Reddit posts",
		UsageText:              "submonkey [OPTIONS..] OUTFILE",
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
			&cli.StringFlag{Name: "size", Aliases: []string{"S"}, Value: "640x360", Usage: "dimensions of the output video"},
		},
	}
}

func execCreate(c *cli.Context) error {
	filter := c.String("filter")
	sort := c.String("sort")
	time := c.String("time")
	limit := c.Int("limit")
	size := strings.ReplaceAll(c.String("size"), "x", ":")
	if c.NArg() < 1 {
		return errors.New("missing output file, see --help for usage details")
	} else if limit > 100 {
		return errors.New("limits must be <= 100")
	} else if !util.InStringList([]string{"hot", "top", "rising", "new", "controversial"}, sort) {
		return errors.New("sort must be any of: hot, top, rising, new, controversial")
	} else if !util.InStringList([]string{"hour", "day", "week", "month", "year", "all"}, time) {
		return errors.New("time must be any of: hour, day, week, month, year, all")
	}
	filename := c.Args().Get(0)
	return submonkey.CreateVideo(&submonkey.CreateOptions{
		Filter:     filter,
		Sort:       sort,
		Time:       time,
		Limit:      limit,
		OutputSize: size,
		OutputFile: filename,
	})
}
