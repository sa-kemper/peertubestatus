package StatsIO

import (
	"strings"
)

// VideoNameToFilePath resolves a video name in order to be used for file operations, since the title may be incompatible with the filesystem e.g. slash in the name.
func VideoNameToFilePath(videoName string) string {
	videoName = strings.ReplaceAll(videoName, "/", `\/`)
	videoName = strings.ToLower(videoName)
	// Why bother making a function for such a small operation? Maintainability. change it once, it breaks everywhere, reliably!
	return videoName
}
