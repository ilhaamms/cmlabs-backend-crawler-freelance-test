package helper

import (
	"net/url"
	"strings"
)

func SanitizeURLToFilename(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		replacer := strings.NewReplacer(
			"https://", "",
			"http://", "",
			"/", "_",
			":", "_",
			"?", "_",
			"&", "_",
			"=", "_",
			".", "_",
		)
		return replacer.Replace(rawURL)
	}

	host := parsed.Hostname()
	host = strings.TrimPrefix(host, "www.")

	replacer := strings.NewReplacer(
		".", "_",
		"-", "_",
	)
	filename := replacer.Replace(host)

	path := strings.Trim(parsed.Path, "/")
	if path != "" {
		pathClean := strings.NewReplacer(
			"/", "_",
			".", "_",
			"-", "_",
		).Replace(path)
		filename = filename + "_" + pathClean
	}

	return filename
}

func PanicIfError(err error) {
	if err != nil {
		panic(err)
	}
}

func ValidateURL(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}

func EnsureScheme(rawURL string) string {
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		return "https://" + rawURL
	}
	return rawURL
}
