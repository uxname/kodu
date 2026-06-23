package fsutil

import "strings"

// MaxFileSizeBytes — порог пропуска крупных файлов (constants.ts:1).
const MaxFileSizeBytes int64 = 1024 * 1024 // 1 MiB

// BinaryProbeSize — сколько байт читать для детекции бинарника по содержимому.
const binaryProbeSize = 8192

// binaryExtensions — расширения, считающиеся бинарными (constants.ts BINARY_EXTENSION_LIST).
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

// knownTextExtensions — заведомо текстовые расширения/имена (constants.ts KNOWN_TEXT_LIST).
// Имеют приоритет над binaryExtensions: совпадение форсирует «текст».
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
