package peertubeApi

import (
	"errors"
)

// VideoIncludeFlags represents the flags for including additional videos in results.
// VideoIncludeFlags are only available from admin tokens.
// Usage example:
//
//	// Combine multiple flags
//	combinedFlags := CombineVideoIncludeFlags(
//	    VideoIncludeFiles,
//	    VideoIncludeCaptions
//	)
//
//	// Validate the flags
//	validFlags, err := ValidateVideoIncludeFlags(combinedFlags, isAdmin)
//	if err != nil {
//	    // Handle error
//	    log.Printf("Flag validation error: %v", err)
//	    return
//	}
//
//	// Use the validated flags in your API call
//	apiCall(validFlags)
//
// Flags represent different video inclusion states:
//   - 0: NONE
//   - 1: NOT_PUBLISHED_STATE
//   - 2: BLACKLISTED
//   - 4: BLOCKED_OWNER
//   - 8: FILES
//   - 16: CAPTIONS
//   - 32: VIDEO SOURCE
type VideoIncludeFlags int

// Enum constants for video include flags
const (
	VideoIncludeNone              VideoIncludeFlags = 0
	VideoIncludeNotPublishedState VideoIncludeFlags = 1
	VideoIncludeBlacklisted       VideoIncludeFlags = 2
	VideoIncludeBlockedOwner      VideoIncludeFlags = 4
	VideoIncludeFiles             VideoIncludeFlags = 8
	VideoIncludeCaptions          VideoIncludeFlags = 16
	VideoIncludeVideoSource       VideoIncludeFlags = 32
)

// ValidateVideoIncludeFlags checks if the provided flags are valid and if the user has permission.
//
// It ensures that:
// 1. Only administrators and moderators can use the parameter
// 2. Only valid flag combinations are allowed
//
// Returns the validated flags or an error if validation fails.
func ValidateVideoIncludeFlags(flags VideoIncludeFlags, isAdminOrModerator bool) (VideoIncludeFlags, error) {
	// Check if the user has administrative privileges
	if !isAdminOrModerator {
		return 0, errors.New("only administrators and moderators can use this parameter")
	}

	// Validate that only allowed flag combinations are used
	allowedFlags := VideoIncludeNone |
		VideoIncludeNotPublishedState |
		VideoIncludeBlacklisted |
		VideoIncludeBlockedOwner |
		VideoIncludeFiles |
		VideoIncludeCaptions |
		VideoIncludeVideoSource

	// Check if the provided flags are a valid combination
	if flags & ^allowedFlags != 0 {
		return 0, errors.New("invalid video include flags")
	}

	return flags, nil
}

// CombineVideoIncludeFlags allows combining multiple flags using bitwise OR.
//
// This method enables creating complex flag combinations by merging individual flags.
// Example:
//
//	combinedFlags := CombineVideoIncludeFlags(VideoIncludeFiles, VideoIncludeCaptions)
func CombineVideoIncludeFlags(flags ...VideoIncludeFlags) VideoIncludeFlags {
	var combinedFlags VideoIncludeFlags
	for _, flag := range flags {
		combinedFlags |= flag
	}
	return combinedFlags
}
