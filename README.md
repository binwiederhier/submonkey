# submonkey 🙈

submonkey is a tool to generate videos from Reddit. It downloads videos from one or more subreddits, and concatenates
and scales them. Under the hood, it uses the powers of [youtube-dl](https://youtube-dl.org/) and [FFmpeg](https://ffmpeg.org/)

I made this just for fun, mainly so I can generate an endless stream of animal videos for my son from 
[r/AnimalsBeingBros](https://www.reddit.com/r/AnimalsBeingBros) and [r/AnimalsBeingDerps](https://www.reddit.com/r/AnimalsBeingDerps).

## Usage
After [installing submonkey](#installation), you may run it like this.

_Example 1:_ Generate a video from the 5 _hottest_ posts on [r/AnimalsBeingBros](https://www.reddit.com/r/AnimalsBeingBros) and 
[r/AnimalsBeingDerps](https://www.reddit.com/r/AnimalsBeingDerps), within the last week, as 360p video (bad quality):
```
$ submonkey animals.mp4
2021/09/13 09:42:09 Retrieving hot posts for subreddit(s) AnimalsBeingBros+AnimalsBeingDerps ...
2021/09/13 09:42:11 Downloading up to 5 video(s) ...
2021/09/13 09:42:12 - Downloaded pmzcar (1/5), https://v.redd.it/myh6xa96j4n71 ...
2021/09/13 09:42:14 - Downloaded pn3o1m (2/5), https://v.redd.it/53zirs95p5n71 ...
2021/09/13 09:42:16 - Downloaded pndyvs (3/5), https://v.redd.it/09lxtrl1g9n71 ...
2021/09/13 09:42:18 - Downloaded pnczrp (4/5), https://v.redd.it/xe2lcvc739n71 ...
2021/09/13 09:42:21 - Downloaded pmv292 (5/5), https://v.redd.it/ewiopd5ie3n71 ...
2021/09/13 09:42:21 Generating video animals.mp4 ...
2021/09/13 09:42:40 Done.
```

_Example 2:_ Generate a video from the 3 _top_ posts on [r/funny](https://www.reddit.com/r/funny), within the last 24 hours, 
as 720p video:
```
$ submonkey \
  --sort top \
  --time day \
  --filter funny \
  --limit 3 \
  --size 720p \
  funny.mp4 
2021/09/13 09:43:07 Retrieving top posts for subreddit(s) funny ...
2021/09/13 09:43:08 Downloading up to 3 video(s) ...
2021/09/13 09:43:09 - Downloaded pn04w9 (1/3), https://v.redd.it/u6gakr7oq4n71 ...
2021/09/13 09:43:11 - Downloaded pmti5s (2/3), https://v.redd.it/jnwu323nz2n71 ...
2021/09/13 09:43:13 - Downloaded pmtuez (3/3), https://i.redd.it/pyjom9l933n71.gif ...
2021/09/13 09:43:13 Generating video funny.mp4 ...
2021/09/13 09:44:05 Done.
```

## Installation
Before installing submonkey, please install [youtube-dl](https://youtube-dl.org/) and [FFmpeg](https://ffmpeg.org/) first.
If you're on Linux, you only need to install `youtube-dl` manually. `ffmpeg` is installed as a package dependency.

Binaries can be found on the [releases page](https://github.com/binwiederhier/submonkey/releases). 

**Debian/Ubuntu** (*from a repository*)**:**   
```bash
# Dependencies: youtube-dl doesn't have a package in the repos
sudo curl -L https://yt-dl.org/downloads/latest/youtube-dl -o /usr/bin/youtube-dl

# Add repository and install submonkey
curl -sSL https://archive.heckel.io/apt/pubkey.txt | sudo apt-key add -
sudo apt install apt-transport-https
sudo sh -c "echo 'deb [arch=amd64] https://archive.heckel.io/apt debian main' > /etc/apt/sources.list.d/archive.heckel.io.list"  
sudo apt update
sudo apt install submonkey
```

**Debian/Ubuntu** (*manual install*)**:**
```bash
# Dependencies: youtube-dl doesn't have a package in the repos
sudo curl -L https://yt-dl.org/downloads/latest/youtube-dl -o /usr/bin/youtube-dl

# Install submonkey from deb file, and ffmpeg
wget https://github.com/binwiederhier/submonkey/releases/download/v0.1.0/submonkey_0.1.0_amd64.deb
dpkg -i submonkey_0.1.0_amd64.deb
apt-get install -f
```

**Fedora/RHEL/CentOS:**
```bash
# Dependencies: youtube-dl doesn't have a package in the repos
sudo curl -L https://yt-dl.org/downloads/latest/youtube-dl -o /usr/bin/youtube-dl

# Install submonkey from rpm file
rpm -ivh https://github.com/binwiederhier/submonkey/releases/download/v0.1.0/submonkey_0.1.0_amd64.rpm
```

**Docker:**   
Since submonkey generates files, you should pass `-u` to ensure that you run it with your own user and group, and 
to mount the `/submonkey` directory wherever you want the output files to be using `-v`. You may also mount 
`/.cache/submonkey` if you like to keep the download cache around.

Please note that the performance of encoding videos using FFmpeg inside of Docker is **orders of magnitudes slower**,
likely due to the fact that you cannot use the GPU.

```bash
mkdir output
docker run \
  -u "$(id -u):$(id -g)" \
  -v "$(pwd)/output:/submonkey" \
  -it binwiederhier/submonkey \
  video.mp4
```

**Go:**
```bash
# Dependencies: youtube-dl doesn't have a package in the repos
sudo curl -L https://yt-dl.org/downloads/latest/youtube-dl -o /usr/bin/youtube-dl
apt-get install ffmpeg # or similar ...

# Get submonkey
go get -u heckel.io/submonkey
```

**Manual install** (*any x86_64-based Linux*)**:**
```bash
wget https://github.com/binwiederhier/submonkey/releases/download/v0.1.0/submonkey_0.1.0_linux_x86_64.tar.gz
sudo tar -C /usr/bin -zxf submonkey_0.1.0_linux_x86_64.tar.gz submonkey
```

## Command-line help
```bash 
$ submonkey --help
NAME:
   submonkey - Create videos from Reddit posts

USAGE:
   submonkey [OPTIONS..] OUTFILE.mp4

GLOBAL OPTIONS:
   --filter value, -f value      source subreddit(s) to be used for videos (default: "AnimalsBeingBros+AnimalsBeingDerps")
   --sort value, -s value        sort posts [hot, top, rising, new, controversial] (default: "hot")
   --time value, -t value        time period to sort posts by [hour, day, week, month, year, all] (default: "week")
   --limit value, -l value       number of posts to include in the video (default: 5)
   --nsfw, -n                    include NSFW content (default: false)
   --size value, -S value        dimensions of the output video [360p, 720p, 1080p, WxH] (default: "360p")
   --cache-dir value, -C value   cache directory for video downloads (default: "/home/pheckel/.cache/submonkey")
   --cache-keep value, -K value  duration after which to delete cache entries (default: "1d")

Try 'submonkey COMMAND --help' for more information.

submonkey 0.1.0 (930324d), runtime go1.17, built at 2021-09-13T13:35:10Z
Copyright (C) 2021 Philipp C. Heckel, distributed under the Apache License 2.0
``` 

## Building
```
make build-simple
# Builds to dist/submonkey_linux_amd64/submonkey
``` 

To build releases, I use [GoReleaser](https://goreleaser.com/). If you have that installed, you can run `make build` or 
`make build-snapshot`.

## Contributing
I welcome any and all contributions. Just create a PR or an issue.

## License
Made with ❤️ by [Philipp C. Heckel](https://heckel.io), distributed under the [Apache License 2.0](LICENSE).

Third party libraries:
* [github.com/urfave/cli/v2](https://github.com/urfave/cli/v2) (MIT) is used to drive the CLI
* [go-reddit](https://github.com/vartanbeno/go-reddit) (MIT) is used to talk to Reddit
* [FFmpeg](https://ffmpeg.org/) (LGPL 2.1) is used to generate videos
* [youtube-dl](https://ffmpeg.org/) (LGPL 2.1) is used to generate videos
* [GoReleaser](https://goreleaser.com/) (MIT) is used to create releases 

Code and posts that helped:
* [FFmpeg filtering introduction](https://ffmpeg.org/ffmpeg-filters.html#Filtering-Introduction)
* [How to concat videos of different sizes](https://stackoverflow.com/a/48853654/1440785)
* [How to use anullsrc to add a silent audio track](https://stackoverflow.com/questions/46057412/ffmpeg-concat-multiple-videos-some-with-audio-some-without/46058429#46058429)
* [Figure out if video has audio](https://stackoverflow.com/a/21447100/1440785)
