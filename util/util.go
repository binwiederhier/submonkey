package util

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

var (
	durationStrSecondsOnlyRegex    = regexp.MustCompile(`(?i)^(\d+)$`)
	durationStrLongPeriodOnlyRegex = regexp.MustCompile(`(?i)^(\d+)([dwy]|mo)$`)
)

// InStringList returns true if needle is contained in the list of strings
func InStringList(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

// FileExists returns true if a file with the given filename exists
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// Run is a shortcut running an exec.Command
func Run(command ...string) error {
	cmd := exec.Command(command[0], command[1:]...)
	return cmd.Run()
}

// ParseDuration is a wrapper around Go's time.ParseDuration to supports days, weeks, months and years ("2y")
// and values without any unit ("1234"), which are interpreted as seconds. This is obviously inaccurate,
// but enough for the use case. In this function, the units are defined as follows:
// - day = 24 hours
// - week = 7 days
// - month = 30 days
// - year = 365 days
func ParseDuration(s string) (time.Duration, error) {
	matches := durationStrSecondsOnlyRegex.FindStringSubmatch(s)
	if matches != nil {
		seconds, err := strconv.Atoi(matches[1])
		if err != nil {
			return -1, fmt.Errorf("cannot convert number %s", matches[1])
		}
		return time.Duration(seconds) * time.Second, nil
	}
	matches = durationStrLongPeriodOnlyRegex.FindStringSubmatch(s)
	if matches != nil {
		number, err := strconv.Atoi(matches[1])
		if err != nil {
			return -1, fmt.Errorf("cannot convert number %s", matches[1])
		}
		switch unit := matches[2]; unit {
		case "d":
			return time.Duration(number) * 24 * time.Hour, nil
		case "w":
			return time.Duration(number) * 7 * 24 * time.Hour, nil
		case "mo":
			return time.Duration(number) * 30 * 24 * time.Hour, nil
		case "y":
			return time.Duration(number) * 365 * 24 * time.Hour, nil
		default:
			return -1, fmt.Errorf("unexpected unit %s", unit)
		}
	}
	return time.ParseDuration(s)
}
