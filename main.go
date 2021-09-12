package main

import (
	"bytes"
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
)

func main() {
	dir := "tmp"
	subreddit := "animalsbeingbros"
	outputScale := "640:360"
	outputFilename := "tmp/output.mp4"
	if err := run(dir, subreddit, outputScale, outputFilename); err != nil {
		log.Fatal(err)
	}
}

func run(dir, subreddit, outputScale, outputFilename string) error {
	ctx := context.Background()
	client := reddit.DefaultClient()

	// Let's get the top 200 posts of r/golang.
	// Reddit returns a maximum of 100 posts at a time,
	// so we'll need to separate this into 2 requests.
	posts, _, err := client.Subreddit.TopPosts(ctx, subreddit, &reddit.ListPostOptions{
		ListOptions: reddit.ListOptions{
			Limit: 10,
		},
		Time: "week",
	})
	if err != nil {
		return err
	}

	filenames := make([]string, 0)
	for _, post := range posts {
		if post.URL == "" || strings.HasSuffix(post.URL, ".jpg") {
			continue
		}
		filename, err := downloadPost(dir, post)
		if err != nil {
			return err
		}
		filenames = append(filenames, filename)
	}

	if output, err := concatVideos(filenames, outputScale, outputFilename); err != nil {
		log.Fatal(string(output))
		return err
	}

	log.Printf("Done.")
	/*
		// The After option sets the id of an item that Reddit
		// will use as an anchor point for the returned listing.
		posts, _, err = reddit.DefaultClient().Subreddit.TopPosts(ctx, "golang", &reddit.ListPostOptions{
			ListOptions: reddit.ListOptions{
				Limit: 100,
				After: resp.After,
			},
			Time: "all",
		})
		if err != nil {
			return err
		}

		for _, post := range posts {
			fmt.Println(post.Title)
		}
	*/
	return nil
}

func downloadPost(dir string, post *reddit.Post) (string, error) {
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

// concatVideos concats multiples videos into one using ffmpeg
//
// Read this to get a basic understanding of what's going on here. It's short and easy to understand:
// https://ffmpeg.org/ffmpeg-filters.html#Filtering-Introduction
//
// Used filters and such:
// - concat: concat video/audio streams, see https://ffmpeg.org/ffmpeg-filters.html#concat
// - scale: resize videos to desired size, adding black padding and centering it, see https://ffmpeg.org/ffmpeg-filters.html#scale-1
//          and https://stackoverflow.com/a/48853654/1440785
// - anullsrc: append silent audio for videos without sound, see https://ffmpeg.org/ffmpeg-filters.html#anullsrc
//             and https://stackoverflow.com/a/46058429/1440785
//
// Full example:
//   ffmpeg -y \
//    -i video1.mp4 \
//    -i video2.mkv \
//    -f lavfi -t 0.1 -i anullsrc \
//  -filter_complex "
//    [0:v]scale=1280:720:force_original_aspect_ratio=decrease,pad=1280:720:(ow-iw)/2:(oh-ih)/2,setsar=1[v0];
//    [1:v]scale=1280:720:force_original_aspect_ratio=decrease,pad=1280:720:(ow-iw)/2:(oh-ih)/2,setsar=1[v1];
//    [v0][0:a][v1][2:a]concat=n=2:v=1:a=1[v][a]
//  " \
//  -map "[v]" -map "[a]" \
// output.mp4
//
func concatVideos(filenames []string, scale, outputFilename string) ([]byte, error) {
	log.Printf("Generating video %s ...", outputFilename)
	var filter, concat string
	for i, filename := range filenames {
		filter += fmt.Sprintf("[%d:v]scale=%s:force_original_aspect_ratio=decrease,pad=%s:(ow-iw)/2:(oh-ih)/2,setsar=1[v%d];\n", i, scale, scale, i)
		withAudio, err := hasAudioStream(filename)
		if err != nil {
			return nil, err
		}
		if withAudio {
			concat += fmt.Sprintf("[v%d][%d:a]", i, i)
		} else {
			concat += fmt.Sprintf("[v%d][%d:a]", i, len(filenames)) // anullsrc (silence audio stream)
		}
	}
	filter += fmt.Sprintf("%sconcat=n=%d:v=1:a=1[v][a]", concat, len(filenames))
	args := make([]string, 0)
	args = append(args, "-y")
	for _, filename := range filenames {
		args = append(args, "-i", filename)
	}
	args = append(args, "-f", "lavfi", "-t", "0.1", "-i", "anullsrc")
	args = append(args, "-filter_complex", filter)
	args = append(args, "-map", "[v]", "-map", "[a]", outputFilename)
	cmd := exec.Command("ffmpeg", args...)
	//log.Printf("command: %s", cmd.String())
	return cmd.CombinedOutput()
}

//
// https://stackoverflow.com/a/21447100/1440785
func hasAudioStream(filename string) (bool, error) {
	cmd := exec.Command("ffprobe", "-show_streams", "-select_streams", "a", "-loglevel", "error", filename)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, err
	}
	return len(bytes.TrimSpace(output)) > 0, nil
}
