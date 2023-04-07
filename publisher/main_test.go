package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateEpisodeCmd(t *testing.T) {
	tempDir, err := os.MkdirTemp(os.TempDir(), "test-posts")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	mockEpNumFn := func(url, reEpisodeNumber string) (int, error) {
		return 5, nil
	}
	nowFn = func() time.Time {
		return time.Date(2023, 4, 7, 14, 40, 46, 0, time.UTC)
	}
	req := PrepEpisode{
		PostsLocation: tempDir,
		ReEpisode:     "https://podcast.umputun.com/",
		Editor:        "", // disable editor for tests
	}

	err = createEpisodeCmd(req, mockEpNumFn)
	require.NoError(t, err)

	outfile := filepath.Join(tempDir, "podcast-5.md")
	assert.FileExists(t, outfile)

	content, err := os.ReadFile(outfile) //nolint:gosec
	assert.NoError(t, err)
	t.Log(string(content))

	exp := `+++
title = "UWP - Выпуск 5"
date = "2023-04-07T14:40:46"
categories = ["podcast"]
image = "https://podcast.umputun.com/images/uwp/uwp5.jpg"
filename = "ump_podcast5"
+++

![](https://podcast.umputun.com/images/uwp/uwp5.jpg)

- .
- .
- .
- .
- .
- .
- .
- Вопросы и ответы

[аудио](https://podcast.umputun.com/media/ump_podcast5.mp3)
<audio src="https://podcast.umputun.com/media/ump_podcast5.mp3" preload="none"></audio>

`
	assert.Equal(t, exp, string(content))
}

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
	tmpFile, e := os.CreateTemp(os.TempDir(), "ump_podcast123.mp3")
	assert.NoError(t, e)
	defer os.Remove(tmpFile.Name())

	reEpisodeNumber := `ump_podcast(\d+)\.mp3`

	{
		episodeNumber, err := getEpisodeNumber(tmpFile.Name(), reEpisodeNumber)
		assert.NoError(t, err)
		assert.Equal(t, 123, episodeNumber)
	}

	{
		invalidFilePath := "non_existent_file.mp3"
		_, err := getEpisodeNumber(invalidFilePath, reEpisodeNumber)
		assert.Error(t, err)
		assert.ErrorContains(t, err, fmt.Sprintf("file not found: %s", invalidFilePath))
	}

	{
		invalidRegex := `invalid_regex(\d+)\.mp3`
		_, err := getEpisodeNumber(tmpFile.Name(), invalidRegex)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "invalid file name")
	}
}
