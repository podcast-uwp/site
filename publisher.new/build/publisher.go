package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	"github.com/bogem/id3v2"
	log "github.com/go-pkgz/lgr"
	"github.com/jessevdk/go-flags"
)

type options struct {
	File    string  `long:"file" env:"FILE" required:"true" description:"mp3 file"`
	Mp3Tags Mp3Tags `group:"mp3" namespace:"mp3" env-namespace:"MP3"`
	Deploy  Deploy  `group:"deploy" namespace:"deploy" env-namespace:"DEPLOY"`
	Dbg     bool    `long:"dbg" env:"DEBUG" description:"debug mode"`
}

// Mp3Tags is a set for mp3 tags, used to parse command line as well as input for setMp3Tags
type Mp3Tags struct {
	Title  string `long:"title" env:"TITLE" default:"UWP Выпуск" description:"title"`
	Artist string `long:"artist" env:"ARTIST" default:"Umputun" description:"artist"`
	Album  string `long:"album" env:"ALBUM" default:"Eженедельный подкаст от Umputun" description:"album"`
	Image  string `long:"image" env:"IMAGE" default:"/srv/cover.jpg" description:"image"`
}

// Deploy is a set for deploy, used to parse command line as well as input for deploy
type Deploy struct {
	Playbook        string `long:"playbook" env:"PLAYBOOK" default:"/srv/ansible.yml" description:"ansible playbook"`
	Host            string `long:"host" env:"HOST" default:"podcast.umputun.com" description:"host"`
	Location        string `long:"location" env:"LOCATION" default:"/srv/media" description:"location"`
	DaysKeep        int    `long:"days-keep" env:"DAYS_KEEP" default:"700" description:"days to keep"`
	ArchiveHost     string `long:"archive-host" env:"ARCHIVE_HOST" default:"master.radio-t.com" description:"archive host"`
	ArchiveLocation string `long:"archive-location" env:"ARCHIVE_LOCATION" default:"/data/archive/uwp/media/" description:"archive location"`
}

var revision = "unknown"

func main() {
	var opts options
	log.Printf("uwp publisher - %s", revision)
	if _, err := flags.Parse(&opts); err != nil {
		os.Exit(1)
	}

	setupLog(opts.Dbg)

	if err := setMp3Tags(opts.File, opts.Mp3Tags); err != nil {
		log.Fatalf("[PANIC] %v", err)
	}

	if err := uploadWithAnsible(opts.File, opts.Deploy); err != nil {
		log.Fatalf("[PANIC] %v", err)
	}
	log.Printf("[INFO] completed %q", opts.File)
}

func setupLog(dbg bool) {
	if dbg {
		log.Setup(log.Debug, log.CallerFile, log.Msec, log.LevelBraces)
		return
	}
	log.Setup(log.Msec, log.LevelBraces)
}

func setMp3Tags(file string, tags Mp3Tags) error {
	log.Printf("[INFO] set mp3 tags for %s, %+v", file, tags)
	num, err := getEpisodeNumber(file)
	if err != nil {
		return err
	}

	episodeFile, err := id3v2.Open(file, id3v2.Options{})
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer episodeFile.Close()

	episodeFile.SetDefaultEncoding(id3v2.EncodingUTF8)

	// Set mp3 tags
	episodeFile.SetTitle(fmt.Sprintf("%s %d", tags.Title, num))
	episodeFile.SetArtist(tags.Artist)
	episodeFile.SetAlbum(tags.Album)
	episodeFile.SetYear(time.Now().Format("2006"))
	episodeFile.SetGenre("Podcast")

	if tags.Image != "" {
		// Open image file
		imageFile, err := os.Open(tags.Image)
		if err != nil {
			return fmt.Errorf("error opening image file: %s", err)
		}
		defer imageFile.Close()

		// Get image data
		imageInfo, err := imageFile.Stat()
		if err != nil {
			return fmt.Errorf("error getting image file info: %s", err)
		}
		imageData := make([]byte, imageInfo.Size())
		if _, err := imageFile.Read(imageData); err != nil {
			return fmt.Errorf("error reading image data: %s", err)
		}

		// Add album art to tags
		pic := id3v2.PictureFrame{
			Encoding:    id3v2.EncodingUTF8,
			MimeType:    "image/jpeg",
			PictureType: id3v2.PTFrontCover,
			Description: "Front cover",
			Picture:     imageData,
		}
		episodeFile.AddAttachedPicture(pic)
	}

	if err := episodeFile.Save(); err != nil {
		return fmt.Errorf("error saving ID3 tags: %v", err)
	}

	return nil
}

func uploadWithAnsible(file string, req Deploy) error {
	log.Printf("[INFO] upload with ansible %s, %+v", file, req)

	inventory := fmt.Sprintf("%s,", req.Host)
	extraVars := fmt.Sprintf("local_file_path=%s remote_directory=%s local_file_path=%s "+
		"num_days=%d archive_host=%s archive_location=%s",
		file, req.Location, file, req.DaysKeep, req.ArchiveHost, req.ArchiveLocation)
	log.Printf("[DEBUG] ansible-playbook %s -i %s -e %s", file, inventory, extraVars)

	cmd := exec.Command("ansible-playbook", req.Playbook, "-i", inventory, "--extra-vars", extraVars)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running ansible-playbook command: %v", err)
	}

	return nil
}

func getEpisodeNumber(filePath string) (int, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return 0, fmt.Errorf("file not found: %s", filePath)
	}
	re := regexp.MustCompile(`ump_podcast(\d+)\.mp3`)
	match := re.FindStringSubmatch(filePath)
	if len(match) == 0 {
		return 0, fmt.Errorf("invalid file name")
	}
	return strconv.Atoi(match[1])
}
