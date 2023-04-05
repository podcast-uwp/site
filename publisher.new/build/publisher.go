package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/bogem/id3v2"
	"github.com/go-pkgz/fileutils"
	log "github.com/go-pkgz/lgr"
	"github.com/jessevdk/go-flags"
)

type options struct {
	Mp3Tags     Mp3Tags     `command:"mp3" description:"set mp3 tags"`
	Deploy      Deploy      `command:"deploy" description:"deploy to remote server"`
	PrepEpisode PrepEpisode `command:"prep" description:"prepare new episode"`
	Sleepy      Sleepy      `command:"sleepy" description:"sleepy mode"`
	Dbg         bool        `long:"dbg" env:"DEBUG" description:"debug mode"`
}

// Mp3Tags is a set for mp3 tags, used to parse command line as well as input for setMp3Tags
type Mp3Tags struct {
	File      string `long:"file" env:"FILE" required:"true" description:"mp3 file"`
	Title     string `long:"title" env:"TITLE" default:"UWP Выпуск" description:"title"`
	Artist    string `long:"artist" env:"ARTIST" default:"Umputun" description:"artist"`
	Album     string `long:"album" env:"ALBUM" default:"Eженедельный подкаст от Umputun" description:"album"`
	Image     string `long:"image" env:"IMAGE" default:"/srv/cover.jpg" description:"image"`
	ReEpisode string `long:"re-episode" env:"RE_EPISODE" default:"ump_podcast(\\d+)\\.mp3" description:"episode num regex"`
}

// Deploy is a set for deploy, used to parse command line as well as input for deploy
type Deploy struct {
	File            string `long:"file" env:"FILE" required:"true" description:"mp3 file"`
	Playbook        string `long:"playbook" env:"PLAYBOOK" default:"/srv/ansible.yml" description:"ansible playbook"`
	Host            string `long:"host" env:"HOST" default:"podcast.umputun.com" description:"host"`
	Location        string `long:"location" env:"LOCATION" default:"/srv/media" description:"location"`
	DaysKeep        int    `long:"days-keep" env:"DAYS_KEEP" default:"700" description:"days to keep"`
	ArchiveHost     string `long:"archive-host" env:"ARCHIVE_HOST" default:"archive.rucast.net" description:"archive host"`
	ArchiveLocation string `long:"archive-location" env:"ARCHIVE_LOCATION" default:"/data/archive/uwp/media/" description:"archive location"`
}

// PrepEpisode is a preparation command of new hugo post for the next episode
type PrepEpisode struct {
	ReEpisode     string `long:"re-episode" env:"RE_EPISODE" default:"ump_podcast(\\d+)\\.mp3" description:"episode num regex"`
	PostsLocation string `long:"location" env:"POSTS_LOCATION" default:"./hugo/content/posts" description:"posts location"`
}

// Sleepy is a command to sleep for a while, needed to start a container and copy files from
type Sleepy struct {
	Duration time.Duration `long:"duration" env:"DURATION" default:"10s" description:"sleep duration"`
}

var revision = "v2.0.1"

func main() {
	var opts options
	log.Printf("uwp publisher - %s", revision)
	p := flags.NewParser(&opts, flags.PrintErrors|flags.PassDoubleDash|flags.HelpFlag)
	p.SubcommandsOptional = true
	if _, err := p.Parse(); err != nil {
		if err.(*flags.Error).Type != flags.ErrHelp {
			fmt.Printf("%v", err)
		}
		os.Exit(1)
	}
	setupLog(opts.Dbg)

	if p.Active != nil && p.Command.Find("prep") == p.Active {
		if err := createEpisode(opts.PrepEpisode); err != nil {
			log.Fatalf("[PANIC] %v", err)
		}
		log.Printf("[INFO] completed episode preparation")
		return
	}

	if p.Active != nil && p.Command.Find("mp3") == p.Active {
		if err := setMp3Tags(opts.Mp3Tags.File, opts.Mp3Tags); err != nil {
			log.Fatalf("[PANIC] %v", err)
		}
		log.Printf("[INFO] completed mp3 tags update")
		return
	}

	if p.Active != nil && p.Command.Find("deploy") == p.Active {
		if err := uploadWithAnsible(opts.Deploy.File, opts.Deploy); err != nil {
			log.Fatalf("[PANIC] %v", err)
		}
		log.Printf("[INFO] completed deploy")
		return
	}

	if p.Active != nil && p.Command.Find("sleepy") == p.Active {
		log.Printf("[INFO] sleeping for %v..", opts.Sleepy.Duration)
		time.Sleep(opts.Sleepy.Duration)
		log.Printf("[INFO] woke up")
		return
	}

	log.Printf("[WARN] nothing to do")
	return
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

	num, err := getEpisodeNumber(file, tags.ReEpisode) // get episode number from file name
	if err != nil {
		// failed if file not found or regex failed
		return err
	}

	// id3v2 will try to rename file. This won't work if file mounted directly to container,
	// so we make a temp copy, set tags and copy back
	tmpFile := filepath.Join(os.TempDir(), filepath.Base(file))
	if err := fileutils.CopyFile(file, tmpFile); err != nil {
		return fmt.Errorf("error copying file %s to %s: %v", file, tmpFile, err)
	}

	episodeFile, err := id3v2.Open(tmpFile, id3v2.Options{})
	if err != nil {
		return fmt.Errorf("error opening file %s: %v", tmpFile, err)
	}

	defer func() {
		if err := episodeFile.Close(); err != nil {
			log.Printf("[WARN] error closing file tags: %v", err)
		}
		if err := os.Remove(tmpFile); err != nil {
			log.Printf("[WARN] error removing temp file %s: %v", tmpFile, err)
		}
	}()

	episodeFile.SetDefaultEncoding(id3v2.EncodingUTF8)
	episodeFile.SetTitle(fmt.Sprintf("%s %d", tags.Title, num))
	episodeFile.SetArtist(tags.Artist)
	episodeFile.SetAlbum(tags.Album)
	episodeFile.SetYear(time.Now().Format("2006"))
	episodeFile.SetGenre("Podcast")

	if tags.Image != "" {
		imageFile, err := os.Open(tags.Image)
		if err != nil {
			return fmt.Errorf("error opening image file: %s", err)
		}
		defer imageFile.Close()

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

	// copy tagged file back
	if err := fileutils.CopyFile(tmpFile, file); err != nil {
		return fmt.Errorf("error copying file %s back to %s: %v", tmpFile, file, err)
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
		return fmt.Errorf("error running ansible-playbook command: %w", err)
	}

	return nil
}

func createEpisode(req PrepEpisode) error {
	wr := func(fh *os.File, s string, f ...interface{}) {
		fmt.Fprintf(fh, s, f...)
		fh.WriteString("\n")
	}

	log.Printf("[INFO] create episode in %s", req.PostsLocation)

	num, err := getNextEpisodeNum(req.ReEpisode)
	if err != nil {
		return fmt.Errorf("error getting next episode number: %w", err)
	}
	log.Printf("[INFO] new episode number: %d", num)

	outfile := filepath.Join(req.PostsLocation, fmt.Sprintf("podcast-%d.md", num))
	log.Printf("[INFO] create episode file %s", outfile)

	f, err := os.Create(outfile)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer f.Close()

	wr(f, "+++")
	wr(f, `title = "UWP - Выпуск %d"`, num)
	wr(f, `date = %q`, time.Now().Format("2006-01-02T15:04:05"))
	wr(f, `categories = ["podcast"]`)
	wr(f, `image = "https://podcast.umputun.com/images/uwp/uwp%d.jpg"`, num)
	wr(f, `filename = "ump_podcast%d"`, num)
	wr(f, "+++")
	wr(f, "")
	wr(f, `![](https://podcast.umputun.com/images/uwp/uwp%d.jpg)`, num)
	wr(f, "")
	wr(f, "- \n- \n- \n- \n- \n- \n- ")
	wr(f, "- Вопросы и ответы")
	wr(f, "")
	wr(f, "[аудио](https://podcast.umputun.com/media/ump_podcast%d.mp3)", num)
	wr(f, `<audio src="https://podcast.umputun.com/media/ump_podcast%d.mp3" preload="none"></audio>`, num)

	// Open the post file in Sublime Text editor
	exec.Command("subl", outfile).Start()
	return nil
}

func getEpisodeNumber(filePath, reEpisodeNumber string) (int, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return 0, fmt.Errorf("file not found: %s", filePath)
	}
	re := regexp.MustCompile(reEpisodeNumber)
	match := re.FindStringSubmatch(filePath)
	if len(match) == 0 {
		return 0, fmt.Errorf("invalid file name")
	}
	return strconv.Atoi(match[1])
}

func getNextEpisodeNum(reEpisodeNumber string) (num int, err error) {
	client := http.Client{Timeout: time.Second * 30}
	resp, err := client.Get("https://podcast.umputun.com/")
	if err != nil {
		return 0, fmt.Errorf("error getting uwp page: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("unexpected status code %v", resp.StatusCode)
	}

	var found bool
	re := regexp.MustCompile(reEpisodeNumber)

	// Scan through the response body line by line
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		match := re.FindStringSubmatch(line)
		if len(match) == 0 {
			continue
		}
		num, err = strconv.Atoi(match[1])
		if err != nil {
			return 0, fmt.Errorf("invalid episode number %s: %w", match[1], err)
		}
		found = true
		break
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("error reading response body: %w", err)
	}

	if !found {
		return 0, fmt.Errorf("ump_podcast not found")
	}

	return num + 1, nil
}
