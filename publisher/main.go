// Package main handles all the uwp preparation and deployment
package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"text/template"
	"time"

	"github.com/bogem/id3v2"
	log "github.com/go-pkgz/lgr"
	"github.com/jessevdk/go-flags"
	"golang.org/x/crypto/ssh"
)

type options struct {
	Mp3Tags     Mp3Tags     `command:"mp3" description:"set mp3 tags"`
	Deploy      Deploy      `command:"deploy" description:"deploy to remote server"`
	PrepEpisode PrepEpisode `command:"prep" description:"prepare new episode"`
	Git         Git         `command:"git" description:"commit and push new episode"`
	Dbg         bool        `long:"dbg" env:"DEBUG" description:"debug mode"`
}

// Mp3Tags is a set for mp3 tags, used to parse command line as well as input for setMp3Tags
type Mp3Tags struct {
	File      string `short:"f" long:"file" env:"FILE" required:"true" description:"mp3 file"`
	Title     string `long:"title" env:"TITLE" default:"UWP Выпуск" description:"title"`
	Artist    string `long:"artist" env:"ARTIST" default:"Umputun" description:"artist"`
	Album     string `long:"album" env:"ALBUM" default:"Eженедельный подкаст от Umputun" description:"album"`
	Image     string `long:"image" env:"IMAGE" default:"" description:"image"`
	ReEpisode string `long:"re-episode" env:"RE_EPISODE" default:"ump_podcast(\\d+)\\.mp3" description:"episode num regex"`
}

// Deploy is a set for deploy, used to parse command line as well as input for deploy
type Deploy struct {
	File            string `short:"f" long:"file" required:"true" description:"mp3 file"`
	Host            string `long:"host" default:"podcast.umputun.com" description:"primary remote host"`
	User            string `long:"user" default:"umputun" description:"remote user"`
	Location        string `long:"location"  default:"/srv/podcast-uwp/var/media" description:"location"`
	DaysKeep        int    `long:"days-keep"  default:"700" description:"days to keep"`
	ArchiveHost     string `long:"archive-host"  default:"archive.rucast.net" description:"archive host"`
	ArchiveLocation string `long:"archive-location"  default:"/data/archive/uwp/media/" description:"archive location"`
	PrivateKeyPath  string `long:"key"  default:"/Users/umputun/.ssh/id_rsa" description:"private key path"`
}

// PrepEpisode is a preparation command of new hugo post for the next episode
type PrepEpisode struct {
	ReEpisode     string `long:"re-episode" env:"RE_EPISODE" default:"ump_podcast(\\d+)\\.mp3" description:"episode num regex"`
	PostsLocation string `long:"location" env:"POSTS_LOCATION" default:"/Users/umputun/dev.umputun/podcast-uwp/hugo/content/posts" description:"posts location"`
	Editor        string `long:"editor" default:"subl" description:"editor"`
}

// Git command commits and pushes changes to the repo
type Git struct {
	Location string `long:"location" default:"/Users/umputun/dev.umputun/podcast-uwp" description:"repo location"`
}

var (
	//go:embed cover.jpg
	imgData []byte

	//go:embed uwp-podcast.tmpl
	tmplData string
)

var nowFn = time.Now // for testing, to override time.Now

var revision = "v2.1.2"

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
	st := time.Now()

	if p.Active != nil && p.Command.Find("prep") == p.Active {
		if err := createEpisodeCmd(opts.PrepEpisode, getEpisodeNumber); err != nil {
			log.Fatalf("[PANIC] %v", err)
		}
		log.Printf("[INFO] completed episode preparation in %v", time.Since(st))
		return
	}

	if p.Active != nil && p.Command.Find("mp3") == p.Active {
		if err := setMp3TagsCmd(opts.Mp3Tags); err != nil {
			log.Fatalf("[PANIC] %v", err)
		}
		log.Printf("[INFO] completed mp3 tags update in %v", time.Since(st))
		return
	}

	if p.Active != nil && p.Command.Find("deploy") == p.Active {
		if err := deployCmd(opts.Deploy); err != nil {
			log.Fatalf("[PANIC] %v", err)
		}
		log.Printf("[INFO] completed deploy in %v", time.Since(st))
		return
	}

	if p.Active != nil && p.Command.Find("git") == p.Active {
		if err := gitCmd(opts.Git); err != nil {
			log.Fatalf("[PANIC] %v", err)
		}
		log.Printf("[INFO] completed git in %v", time.Since(st))
		return
	}

	log.Printf("[WARN] nothing to do")
}

// setMp3TagsCmd sets mp3 tags for the given file. By default it uses the embedded cover image,
// but can be overridden with --image flag
func setMp3TagsCmd(req Mp3Tags) error {
	log.Printf("[INFO] set mp3 tags for %+v", req)

	num, err := getEpisodeNumber(req.File, req.ReEpisode) // get episode number from file name
	if err != nil {
		// failed if file not found or regex failed
		return fmt.Errorf("error getting episode number from %s: %w", req.File, err)
	}

	origFinfo, err := os.Stat(req.File)
	if err != nil {
		return fmt.Errorf("error getting file info %s: %w", req.File, err)
	}
	log.Printf("[DEBUG] file info for %s - time: %s, size: %d",
		req.File, origFinfo.ModTime().Format(time.RFC3339), origFinfo.Size())

	episodeFile, err := id3v2.Open(req.File, id3v2.Options{})
	if err != nil {
		return fmt.Errorf("error opening file %s: %w", req.File, err)
	}

	defer func() {
		if err := episodeFile.Close(); err != nil {
			log.Printf("[WARN] error closing file tags: %v", err)
		}
		if err := os.Chtimes(req.File, origFinfo.ModTime(), origFinfo.ModTime()); err != nil {
			log.Printf("[WARN] error setting file time %s %s: %v", req.File, origFinfo.ModTime().Format(time.RFC3339), err)
		}
	}()

	episodeFile.SetDefaultEncoding(id3v2.EncodingUTF8)
	episodeFile.SetTitle(fmt.Sprintf("%s %d", req.Title, num))
	episodeFile.SetArtist(req.Artist)
	episodeFile.SetAlbum(req.Album)
	episodeFile.SetYear(origFinfo.ModTime().Format("2006"))
	episodeFile.SetGenre("Podcast")

	if req.Image != "" {
		imageFile, err := os.Open(req.Image)
		if err != nil {
			return fmt.Errorf("error opening image file: %s", err)
		}
		defer imageFile.Close() // nolint

		imageInfo, err := imageFile.Stat()
		if err != nil {
			return fmt.Errorf("error getting image file info: %s", err)
		}
		imgData = make([]byte, imageInfo.Size())
		if _, err := imageFile.Read(imgData); err != nil {
			return fmt.Errorf("error reading image data: %s", err)
		}
	}

	// Add album art to tags
	pic := id3v2.PictureFrame{
		Encoding:    id3v2.EncodingUTF8,
		MimeType:    "image/jpeg",
		PictureType: id3v2.PTFrontCover,
		Description: "Front cover",
		Picture:     imgData,
	}
	episodeFile.AddAttachedPicture(pic)

	if err := episodeFile.Save(); err != nil {
		return fmt.Errorf("error saving ID3 tags: %v", err)
	}

	return nil
}

// createEpisodeCmd makes a new hugo post for the next episode
func createEpisodeCmd(req PrepEpisode, epNumFn func(url, reEpisodeNumber string) (int, error)) error {
	log.Printf("[INFO] create episode in %s", req.PostsLocation)

	num, err := epNumFn("https://podcast.umputun.com/", req.ReEpisode)
	if err != nil {
		return fmt.Errorf("error getting next episode number: %w", err)
	}
	log.Printf("[INFO] new episode number: %d", num)

	outfile := filepath.Join(req.PostsLocation, fmt.Sprintf("podcast-%d.md", num))
	log.Printf("[INFO] create episode file %s", outfile)
	if err = os.MkdirAll(req.PostsLocation, 0o750); err != nil {
		return fmt.Errorf("error creating posts dir %s: %w", req.PostsLocation, err)
	}

	f, err := os.Create(outfile) //nolint:gosec
	if err != nil {
		return fmt.Errorf("error creating file %s: %w", outfile, err)
	}
	defer f.Close() // nolint

	data := struct {
		Number int
		Date   string
	}{
		Number: num,
		Date:   nowFn().Format("2006-01-02T15:04:05"),
	}

	tmpl, err := template.New("episode").Parse(tmplData)
	if err != nil {
		return fmt.Errorf("error parsing template: %w", err)
	}

	err = tmpl.Execute(f, data)
	if err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	if err = f.Sync(); err != nil {
		return fmt.Errorf("error syncing file %s: %w", f.Name(), err)
	}

	// Open the post file in text editor if specified
	if req.Editor != "" {
		if err = exec.Command(req.Editor, outfile).Start(); err != nil { //nolint:gosec
			return fmt.Errorf("error opening file in editor %q: %w", req.Editor, err)
		}
	}
	return nil
}

// deploy uploads file to podcast server and to archive server.
// it also removes old files from podcast server.
func deployCmd(req Deploy) error {
	log.Printf("[INFO] deploy %+v", req)

	key, err := os.ReadFile(req.PrivateKeyPath)
	if err != nil {
		return fmt.Errorf("unable to read private key: %v", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return fmt.Errorf("unable to parse private key: %v", err)
	}

	sshConfig := &ssh.ClientConfig{User: req.User, Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey()} // nolint

	// create remote directory
	if err = sshRun(sshConfig, req.Host, fmt.Sprintf("mkdir -p %s", req.Location)); err != nil {
		return fmt.Errorf("error creating remote directory: %v", err)
	}

	// copy file to remote server
	if err = scpUpload(sshConfig, req.File, req.Host, req.Location, req.PrivateKeyPath); err != nil {
		return fmt.Errorf("error copying file to remote server: %v", err)
	}

	// get list of old files on remote server and delete them
	oldFilesCmd := fmt.Sprintf("find %s -type f -name '*.mp3' -mtime +%d -exec rm -f {} \\;", req.Location, req.DaysKeep)
	if err = sshRun(sshConfig, req.Host, oldFilesCmd); err != nil {
		return fmt.Errorf("error deleting old files on remote server: %v", err)
	}

	// create archive directory on archive server
	if err = sshRun(sshConfig, req.ArchiveHost, fmt.Sprintf("mkdir -p %s", req.ArchiveLocation)); err != nil {
		return fmt.Errorf("error creating archive directory on archive server: %v", err)
	}

	// copy file to archive server
	if err = scpUpload(sshConfig, req.File, req.ArchiveHost, req.ArchiveLocation, req.PrivateKeyPath); err != nil {
		return fmt.Errorf("error copying file to archive server: %v", err)
	}

	return nil
}

// gitCmd pulls changes from git repo, checks for changes and commits them if any. It also pushes changes to remote.
func gitCmd(req Git) error {
	cmd := exec.Command("git", "pull")
	cmd.Dir, cmd.Stdout, cmd.Stderr = req.Location, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error pulling changes: %v", err)
	}

	cmd = exec.Command("git", "status", "--porcelain")
	cmd.Dir = req.Location
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error getting git status: %v", err)
	}

	if len(output) == 0 {
		log.Printf("[INFO] no changes found")
		return nil
	}

	log.Printf("[INFO] changes found")
	fmt.Println(string(output))

	cmd = exec.Command("git", "add", ".")
	cmd.Dir, cmd.Stdout, cmd.Stderr = req.Location, os.Stdout, os.Stderr
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("error adding changes: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", fmt.Sprintf("auto-update %s", time.Now().Format("2006-01-02 15:04:05"))) //nolint:gosec
	cmd.Dir, cmd.Stdout, cmd.Stderr = req.Location, os.Stdout, os.Stderr
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("error committing changes: %v", err)
	}

	cmd = exec.Command("git", "push")
	cmd.Dir, cmd.Stdout, cmd.Stderr = req.Location, os.Stdout, os.Stderr
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("error pushing changes: %v", err)
	}
	return nil
}

func sshRun(sshConfig *ssh.ClientConfig, host, command string) error {
	log.Printf("[DEBUG] run command %q on %s", command, host)
	client, err := ssh.Dial("tcp", host+":22", sshConfig)
	if err != nil {
		return fmt.Errorf("failed to dial: %v", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	session.Stdout, session.Stderr = os.Stdout, os.Stderr
	err = session.Run(command)
	if err != nil {
		return fmt.Errorf("failed to run command: %v", err)
	}
	return nil
}

func scpUpload(sshConfig *ssh.ClientConfig, localFile, host, remoteDir, key string) error {
	log.Printf("[INFO] upload %s to %s:%s", localFile, host, remoteDir)
	defer func(st time.Time) { log.Printf("[DEBUG] upload done in %s", time.Since(st)) }(time.Now())

	remotePath := fmt.Sprintf("%s@%s:%s", sshConfig.User, host, remoteDir)
	cmd := exec.Command("scp", "-i", key, "-o", "StrictHostKeyChecking=no", localFile, remotePath) //nolint:gosec
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run SCP command: %v", err)
	}
	return nil
}

// getEpisodeNumber returns episode number from file name
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

// getNextEpisodeNum returns next episode number by parsing uwp page
func getNextEpisodeNum(url, reEpisodeNumber string) (int, error) {
	client := http.Client{Timeout: time.Second * 30}
	resp, err := client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("error getting uwp page: %w", err)
	}
	defer resp.Body.Close() //nolint:gosec

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("invalid status code %d", resp.StatusCode)
	}

	re := regexp.MustCompile(reEpisodeNumber)
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		line := scanner.Text()
		match := re.FindStringSubmatch(line)
		if len(match) > 0 {
			num, err := strconv.Atoi(match[1])
			if err != nil {
				return 0, fmt.Errorf("invalid episode number %s: %w", match[1], err)
			}
			log.Printf("[DEBUG] found episode %d in %s ", num, line)
			return num + 1, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("error reading response body: %w", err)
	}

	return 0, fmt.Errorf("ump_podcast not found")
}

func setupLog(dbg bool) {
	if dbg {
		log.Setup(log.Debug, log.CallerFile, log.Msec, log.LevelBraces)
		return
	}
	log.Setup(log.Msec, log.LevelBraces)
}
