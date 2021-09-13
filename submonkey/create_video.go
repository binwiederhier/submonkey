package submonkey

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/vartanbeno/go-reddit/v2/reddit"
	"io/ioutil"
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
	filenames, err := downloadFiles(opts, posts)
	if err != nil {
		return err
	}
	log.Printf("Generating video %s ...", opts.OutputFile)
	if output, err := util.ConcatVideos(filenames, opts.OutputSize, opts.OutputFile); err != nil {
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

func downloadFiles(opts *CreateOptions, posts []*reddit.Post) ([]string, error) {
	if err := os.MkdirAll(opts.CacheDir, 0700); err != nil {
		return nil, err
	}
	filenames := make([]string, 0)
	for _, post := range posts {
		if err := includePost(opts, post); err != nil {
			continue
		}
		filename, err := downloadFile(opts.CacheDir, post)
		if err != nil {
			continue
		}
		filenames = append(filenames, filename)
		if len(filenames) == opts.Limit {
			break
		}
	}
	return filenames, nil
}

func downloadFile(cacheDir string, post *reddit.Post) (string, error) {
	imageFilename := filepath.Join(cacheDir, post.ID+".mp4")
	metaFilename := filepath.Join(cacheDir, post.ID+".json")
	if _, err := os.Stat(imageFilename); err == nil {
		log.Printf("Already downloaded %s, %s", post.ID, post.URL)
		return imageFilename, nil
	}
	args := []string{
		"--output", imageFilename,
		"--merge-output-format", "mp4",
		post.URL,
	}
	cmd := exec.Command("youtube-dl", args...)
	if err := cmd.Run(); err != nil {
		return "", err
	}
	metaBytes, err := json.Marshal(post)
	if err != nil {
		return "", err
	}
	if err := ioutil.WriteFile(metaFilename, metaBytes, 0600); err != nil {
		return "", err
	}
	log.Printf("Downloaded %s, %s ...", post.ID, post.URL)
	return imageFilename, nil
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
