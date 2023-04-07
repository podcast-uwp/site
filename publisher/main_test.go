package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNextEpisodeNum(t *testing.T) {
	samplePodcastPage := `
<!DOCTYPE html>
<html>
<head>
	<title>Podcast</title>
</head>
<body>
	<div class="episode">
		<a href="/podcasts/ump_podcast571.mp3">UWP Podcast Episode 571</a>
	</div>
</body>
</html>
`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, samplePodcastPage)
	}))
	defer ts.Close()

	reEpisodeNumber := `ump_podcast(\d+)\.mp3`
	nextEpisode, err := getNextEpisodeNum(ts.URL, reEpisodeNumber)
	assert.NoError(t, err)
	assert.Equal(t, 572, nextEpisode)
}

func TestGetEpisodeNumber(t *testing.T) {
	tmpFile, err := os.CreateTemp(os.TempDir(), "ump_podcast123.mp3")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	reEpisodeNumber := `ump_podcast(\d+)\.mp3`

	{
		episodeNumber, err := getEpisodeNumber(tmpFile.Name(), reEpisodeNumber)
		assert.NoError(t, err)
		assert.Equal(t, 123, episodeNumber)
	}

	{
		invalidFilePath := "non_existent_file.mp3"
		_, err = getEpisodeNumber(invalidFilePath, reEpisodeNumber)
		assert.Error(t, err)
		assert.ErrorContains(t, err, fmt.Sprintf("file not found: %s", invalidFilePath))
	}

	{
		invalidRegex := `invalid_regex(\d+)\.mp3`
		_, err = getEpisodeNumber(tmpFile.Name(), invalidRegex)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "invalid file name")
	}
}
