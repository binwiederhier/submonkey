package util

import (
	"bytes"
	"fmt"
	"os/exec"
)

// ConcatVideos concats multiples videos into one using ffmpeg
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
func ConcatVideos(infiles []string, scale, outfile string) ([]byte, error) {
	var filter, concat string
	for i, filename := range infiles {
		filter += fmt.Sprintf("[%d:v]scale=%s:force_original_aspect_ratio=decrease,pad=%s:(ow-iw)/2:(oh-ih)/2,setsar=1[v%d];\n", i, scale, scale, i)
		withAudio, err := HasAudioStream(filename)
		if err != nil {
			return nil, err
		}
		if withAudio {
			concat += fmt.Sprintf("[v%d][%d:a]", i, i)
		} else {
			concat += fmt.Sprintf("[v%d][%d:a]", i, len(infiles)) // anullsrc (silence audio stream)
		}
	}
	filter += fmt.Sprintf("%sconcat=n=%d:v=1:a=1[v][a]", concat, len(infiles))
	args := make([]string, 0)
	args = append(args, "-y")
	for _, filename := range infiles {
		args = append(args, "-i", filename)
	}
	args = append(args, "-f", "lavfi", "-t", "0.1", "-i", "anullsrc")
	args = append(args, "-filter_complex", filter)
	args = append(args, "-map", "[v]", "-map", "[a]", outfile)
	cmd := exec.Command("ffmpeg", args...)
	return cmd.CombinedOutput()
}

// HasAudioStream returns true if the given file has an audio stream and false if it has not
// See https://stackoverflow.com/a/21447100/1440785
func HasAudioStream(filename string) (bool, error) {
	cmd := exec.Command("ffprobe", "-show_streams", "-select_streams", "a", "-loglevel", "error", filename)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, err
	}
	return len(bytes.TrimSpace(output)) > 0, nil
}
