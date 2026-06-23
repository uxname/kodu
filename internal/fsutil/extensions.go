package fsutil

import "strings"

// MaxFileSizeBytes is the threshold for skipping large files (constants.ts:1).
const MaxFileSizeBytes int64 = 1024 * 1024 // 1 MiB

// BinaryProbeSize is how many bytes to read for content-based binary detection.
const binaryProbeSize = 8192

// binaryExtensions are extensions considered binary (constants.ts BINARY_EXTENSION_LIST).
var binaryExtensions = toSet([]string{
	".png", ".jpg", ".jpeg", ".webp", ".gif", ".bmp", ".ico", ".tif", ".tiff",
	".psd", ".ai", ".sketch", ".heic", ".heif",
	".mp3", ".wav", ".flac", ".ogg", ".m4a",
	".mp4", ".mkv", ".mov", ".avi", ".webm", ".wmv", ".flv", ".mpg", ".mpeg", ".ogv",
	".zip", ".gz", ".tgz", ".bz2", ".xz", ".rar", ".7z", ".tar",
	".pdf", ".exe", ".dll", ".so", ".dylib", ".class", ".jar", ".war", ".ear",
	".ttf", ".otf", ".woff", ".woff2", ".eot",
	".bin", ".pak", ".dat",
})

// knownTextExtensions are extensions/names known to be text (constants.ts KNOWN_TEXT_LIST).
// They take priority over binaryExtensions: a match forces "text".
var knownTextExtensions = toSet([]string{
	// Web / JS
	".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs", ".json", ".html", ".css",
	".scss", ".less", ".vue", ".svelte",
	// Backend / System
	".java", ".py", ".c", ".cpp", ".h", ".hpp", ".cs", ".go", ".rs", ".php",
	".rb", ".swift", ".kt", ".dart", ".scala", ".pl", ".lua", ".sh", ".bat",
	// Data / Docs
	".md", ".txt", ".xml", ".yaml", ".yml", ".sql", ".graphql", ".toml", ".ini", ".env",
	// Config
	".gitignore", ".dockerignore", "dockerfile", ".editorconfig",
})

func toSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, v := range values {
		set[strings.ToLower(v)] = struct{}{}
	}
	return set
}
