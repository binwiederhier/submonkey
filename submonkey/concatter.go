package submonkey

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/vartanbeno/go-reddit/v2/reddit"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"submonkey/util"
)

type CreateOptions struct {
	Filter     string
	Sort       string
	Time       string
	Limit      int
	OutputSize string
	OutputFile string
}

func CreateVideo(opts *CreateOptions) error {
	if err := checkDependencies(); err != nil {
		return err
	}
	log.Printf("Retrieving %s posts for subreddit(s) %s ...", opts.Sort, opts.Filter)
	posts, err := retrievePosts(opts)
	if err != nil {
		return err
	}
	filenames, err := downloadFiles(posts)
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
		Limit: opts.Limit, // max. is 100, otherwise we have to use response.After
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

func downloadFiles(posts []*reddit.Post) ([]string, error) {
	filenames := make([]string, 0)
	for _, post := range posts {
		if post.URL == "" || strings.HasSuffix(post.URL, ".jpg") {
			continue
		}
		filename, err := downloadFile("tmp", post)
		if err != nil {
			return nil, err
		}
		filenames = append(filenames, filename)
	}
	return filenames, nil
}

func downloadFile(dir string, post *reddit.Post) (string, error) {
	imageFilename := filepath.Join(dir, post.ID+".mp4")
	metaFilename := filepath.Join(dir, post.ID+".json")
	if _, err := os.Stat(imageFilename); err == nil {
		log.Printf("Already downloaded %s, %s", post.ID, post.URL)
		return imageFilename, nil
	}
	log.Printf("Downloading %s, %s ...", post.ID, post.URL)
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
	return imageFilename, nil
}
