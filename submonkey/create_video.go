package submonkey

import (
	"context"
	"errors"
	"fmt"
	"github.com/vartanbeno/go-reddit/v2/reddit"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"submonkey/util"
	"time"
)

var (
	regexExcludeExts = regexp.MustCompile(`(png|jpe?g)$`)
)

type CreateOptions struct {
	Filter     string
	Sort       string
	Time       string
	Limit      int
	NSFW       bool
	OutputSize string
	OutputFile string
	CacheDir   string
	CacheKeep  time.Duration
	Version    string
}

func CreateVideo(opts *CreateOptions) error {
	cleanCache(opts.CacheDir, opts.CacheKeep)
	defer cleanCache(opts.CacheDir, opts.CacheKeep)
	if err := checkDependencies(); err != nil {
		return err
	}
	log.Printf("Retrieving %s posts for subreddit(s) %s ...", opts.Sort, opts.Filter)
	posts, err := retrievePosts(opts)
	if err != nil {
		return err
	}
	log.Printf("Downloading up to %d video(s) ...", opts.Limit)
	filenames, posts, err := downloadFiles(opts, posts)
	if err != nil {
		return err
	}
	log.Printf("Generating video %s ...", opts.OutputFile)
	comment := generateComment(posts, opts)
	if output, err := util.ConcatVideos(filenames, opts.OutputSize, comment, opts.OutputFile); err != nil {
		log.Fatal(string(output))
		return err
	}
	log.Printf("Done.")
	return nil
}

func checkDependencies() error {
	if err := util.Run("ffmpeg", "-version"); err != nil {
		return fmt.Errorf("ffmpeg check failed, please install ffmpeg: %s", err.Error())
	} else if err := util.Run("ffprobe", "-version"); err != nil {
		return fmt.Errorf("ffprobe check failed, please install ffmpeg: %s", err.Error())
	} else if err := util.Run("youtube-dl", "--version"); err != nil {
		return fmt.Errorf("youtube-dl check failed, please install youtube-dl: %s", err.Error())
	}
	return nil
}

func retrievePosts(opts *CreateOptions) (posts []*reddit.Post, err error) {
	ctx := context.Background()
	client := reddit.DefaultClient()
	listOpts := &reddit.ListOptions{
		Limit: 100, // always 100, so we can filter out without extra logic; if this is ever > 100, use resp.After
	}
	listPostOpts := &reddit.ListPostOptions{
		ListOptions: *listOpts,
		Time:        opts.Time,
	}
	switch opts.Sort {
	case "hot":
		posts, _, err = client.Subreddit.HotPosts(ctx, opts.Filter, listOpts)
	case "top":
		posts, _, err = client.Subreddit.TopPosts(ctx, opts.Filter, listPostOpts)
	case "rising":
		posts, _, err = client.Subreddit.RisingPosts(ctx, opts.Filter, listOpts)
	case "new":
		posts, _, err = client.Subreddit.NewPosts(ctx, opts.Filter, listOpts)
	case "controversial":
		posts, _, err = client.Subreddit.ControversialPosts(ctx, opts.Filter, listPostOpts)
	default:
		err = fmt.Errorf("invalid sort options: %s", opts.Sort)
	}
	return
}

func downloadFiles(opts *CreateOptions, inposts []*reddit.Post) (filenames []string, posts []*reddit.Post, err error) {
	if err := os.MkdirAll(opts.CacheDir, 0700); err != nil {
		return nil, nil, err
	}
	filenames = make([]string, 0)
	posts = make([]*reddit.Post, 0)
	for _, post := range inposts {
		if err := includePost(opts, post); err != nil {
			continue
		}
		filename, err := downloadFile(opts, post, len(filenames)+1)
		if err != nil {
			continue
		}
		filenames = append(filenames, filename)
		posts = append(posts, post)
		if len(filenames) == opts.Limit {
			break
		}
	}
	if len(filenames) < opts.Limit {
		log.Printf("- No other videos in listing")
	}
	return filenames, posts, nil
}

func downloadFile(opts *CreateOptions, post *reddit.Post, num int) (string, error) {
	videoFile := filepath.Join(opts.CacheDir, post.ID+".mp4")
	if _, err := os.Stat(videoFile); err == nil {
		log.Printf("- Already downloaded %s (%d/%d), %s", post.ID, num, opts.Limit, post.URL)
		return videoFile, nil
	}
	args := []string{
		"--output", videoFile,
		"--merge-output-format", "mp4",
		post.URL,
	}
	cmd := exec.Command("youtube-dl", args...)
	if err := cmd.Run(); err != nil {
		return "", err
	}
	log.Printf("- Downloaded %s (%d/%d), %s ...", post.ID, num, opts.Limit, post.URL)
	return videoFile, nil
}

func includePost(opts *CreateOptions, post *reddit.Post) error {
	if post.URL == "" {
		return errors.New("empty URL")
	} else if post.NSFW && !opts.NSFW {
		return errors.New("tagged NSFW")
	} else if regexExcludeExts.MatchString(post.URL) {
		return errors.New("unsupported file")
	}
	return nil
}

func generateComment(posts []*reddit.Post, opts *CreateOptions) string {
	comment := fmt.Sprintf("Created with submonkey %s, https://heckel.io/submonkey\n", opts.Version)
	comment += fmt.Sprintf("Subreddit(s): %s, sorted by %s (%s)\n\n", opts.Filter, opts.Sort, opts.Time)
	comment += "Videos:\n\n"
	for i, post := range posts {
		comment += fmt.Sprintf("#%d: %s\nhttps://reddit.com%s\n%s\n\n", i+1, post.Title, post.Permalink, post.URL)
	}
	return comment
}

func cleanCache(cacheDir string, keep time.Duration) {
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil || time.Since(info.ModTime()) < keep {
			continue
		}
		_ = os.Remove(filepath.Join(cacheDir, entry.Name()))
	}
}
