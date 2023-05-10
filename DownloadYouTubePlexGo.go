package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type settings struct {
	Email             string
	MediaFolder       string
	Config            string
	PlaylistItems     string
	PushoverUserToken string
	PodcastDownload   []YouTubeDownload `xml:"PodcastDownload"`
}

type Entry struct {
	Title     string
	Published string
	Updated   string
	ID        string
	Link      string
	Content   string
}

type Validate struct {
	MediaFolder                     bool
	Config                          bool
	PodcastDownload_Name            bool
	PodcastDownload_ChannelID       bool
	PodcastDownload_DownloadArchive bool
	PodcastDownload_FileFormat      bool
	PodcastDownload_FileQuality     bool
	PlaylistItems                   bool
	PushoverUserToken               bool
	PodcastDownload_YouTubeURL      bool
}

type YouTubeDownload struct {
	Name             string `xml:"Name"`
	ChannelID        string `xml:"ChannelID"`
	FileFormat       string `xml:"FileFormat"`
	DownloadArchive  string `xml:"DownloadArchive"`
	FileQuality      string `xml:"FileQuality"`
	ChannelThumbnail string `xml:"ChannelThumbnail"`
	YouTubeURL       string `xml:"YouTubeURL"`
	PushoverAppToken string `xml:"PushoverAppToken"`
	// PushoverAppToken
}

type JsonData struct {
	id              string
	title           string
	webpage_url     string
	thumbnail       string
	description     string
	uploader_url    string
	channel_url     string
	duration_string string
	filesize_approx float64
}

type JsonChannelData struct {
	thumbnail   string
	description string
}

func isOlderThan(t time.Time) bool {
	return time.Now().Sub(t) > 168*time.Hour
}

func DeleteOldFiles(dir string) {
	descfiles, descerr := WalkMatch(dir, "*.description")

	if descerr != nil {
		log.Printf("------------------      START List description Files ERROR")
		log.Fatal(descerr)
		log.Printf("------------------      END List description Files ERROR")
	}

	for _, fname := range descfiles {
		arrfname_noext := strings.Split(fname, ".")
		fname_noext := arrfname_noext[0]

		log.Printf("fname:" + fname)
		log.Printf("fname_noext:" + fname_noext)

		fname_file, fname_fileerr := os.Stat(fname)

		if fname_fileerr != nil {
			log.Printf("------------------      START List fname_fileerr ERROR")
			log.Fatal(fname_fileerr)
			log.Printf("------------------      END List fname_fileerr ERROR")
		}

		if isOlderThan(fname_file.ModTime()) {
			log.Printf("DELETE FILE: " + fname_noext + ".description")
			os.Remove(fname_noext + ".description")
			log.Printf("DELETE FILE: " + fname_noext + ".mp4")
			os.Remove(fname_noext + ".mp4")
			log.Printf("DELETE FILE: " + fname_noext + ".info.json")
			os.Remove(fname_noext + ".info.json")
		}
	}
}

func IsValid(fp string) bool {
	// Check if file already exists
	if _, err := os.Stat(fp); err == nil {
		return true
	}
	return false
}

func handleJSONObject(object interface{}, key, indentation string) {
	switch t := object.(type) {
	case string:
		fmt.Println(indentation+key+": ", t) // t has type string
	case bool:
		fmt.Println(indentation+key+": ", t) // t has type bool
	case float64:
		fmt.Println(indentation+key+": ", t) // t has type float64 (which is the type used for all numeric types)
	case map[string]interface{}:
		fmt.Println(indentation + key + ":")
		for k, v := range t {
			handleJSONObject(v, k, indentation+"\t")
		}
	case []interface{}:
		fmt.Println(indentation + key + ":")
		for index, v := range t {
			handleJSONObject(v, "["+strconv.Itoa(index)+"]", indentation+"\t")
		}
	}
}

func IsValidURL(fp string) bool {
	log.Printf("URL Check: " + fp)

	resp, err := http.Get(fp)
	if err != nil {
		// print(err.Error())
		log.Printf("IsValidURL Status: " + resp.Status)
		// log.Printf("IsValidURL Error: " + err.Error())
		return false
	} else {
		if strings.Contains(resp.Status, "200 OK") {
			// print(string(resp.StatusCode) + resp.Status)
			log.Printf("URL Status: " + resp.Status)
			return true
		} else {
			// print(err.Error())
			log.Printf("IsValidURL Status: " + resp.Status)
			return false
		}

	}
}

func createKeyValuePairs(m map[string]string) string {
	b := new(bytes.Buffer)
	for key, value := range m {
		fmt.Fprintf(b, "%s=\"%s\"\n", key, value)
	}
	return b.String()
}

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func WalkMatch(root, pattern string) ([]string, error) {
	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if matched {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}

func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func NotifyPushover(Config string, AppToken string, UserToken string, nTitle string, nBody string, pThumbnail string, nURL string) {
	// NotifyPushover("apb75jkyb1iegxzp4styr5tgidq3fg","RSS Podcast Downloaded (" + pName + ")","<html><body>" + ytvideo_title + "<br /><br />--------------------------------------------<br /><br />" + ytvideo_description + "</body></html>",ytvideo_thumbnail)

	log.Println("-----		")
	log.Println("-----		START NotifyPushover")
	log.Println("-----		")

	// ~~~~~~~~~~~~~~ Print Data ~~~~~~~~~~~~~~~~

	log.Printf("------------------      START Print Data")
	log.Printf("Config: " + Config)
	log.Printf("AppToken: " + AppToken)
	log.Printf("UserToken: " + UserToken)
	log.Printf("nTitle: " + nTitle)
	log.Printf("nBody: " + nBody)
	log.Printf("pThumbnail: " + pThumbnail)
	log.Printf("nURL: " + nURL)
	log.Printf("------------------      END Print Data")

	// ~~~~~~~~~~ Download Thumbnail ~~~~~~~~~~~~

	savename := ""
	if strings.HasSuffix(pThumbnail, ".jpg") {
		savename = "maxresdefault.jpg"
	}

	if strings.HasSuffix(pThumbnail, ".webp") {
		savename = "maxresdefault.webp"
	}

	if strings.HasSuffix(pThumbnail, ".jpg") == false && strings.HasSuffix(pThumbnail, ".webp") == false {
		savename = "maxresdefault.jpg"
	}

	err := DownloadFile(Config+savename, pThumbnail)
	if err != nil {
		// panic(err)
		log.Printf("------------------      START DownloadFile ERROR")
		log.Fatal(err.Error())
		log.Printf("------------------      END DownloadFile ERROR")
	}
	fmt.Println("Downloaded: " + pThumbnail)

	// ~~~~~~~~~~~~~~ HTTP Post ~~~~~~~~~~~~~~~~~

	out := exec.Command("curl", "-s", "--form-string", "token="+AppToken, "--form-string", "user="+UserToken, "--form-string", "title="+nTitle, "--form-string", "message="+nBody, "--form-string", "html=1", "-F", "attachment=@"+Config+savename, "https://api.pushover.net/1/messages.json")
	out.Stdout = os.Stdout
	out.Stderr = os.Stderr

	if err := out.Run(); err != nil {
		log.Printf("------------------      START NotifyPushover ERROR")
		log.Fatal(err.Error())
		log.Printf("------------------      END NotifyPushover ERROR")
	}

	log.Println("-----		END NotifyPushover")
}

func Run_YTDLP(sMediaFolder string, Config string, pName string, pChannelID string, pFileFormat string, pDownloadArchive string, pFileQuality string, PlaylistItems string, pYouTubeURL string, pPushoverAppToken string, pPushoverUserToken string) {
	log.Println("-----		")
	log.Println("-----		Start Run_YTDLP")
	log.Println("-----		")

	// ~~~~~~~~~~~~~~ Print Data ~~~~~~~~~~~~~~~~
	log.Println("-----		")
	log.Printf("sMediaFolder: " + sMediaFolder)
	log.Printf("Config: " + Config)
	log.Printf("pName: " + pName)
	log.Printf("pChannelID: " + pChannelID)
	log.Printf("pFileFormat: " + pFileFormat)
	log.Printf("pDownloadArchive: " + pDownloadArchive)
	log.Printf("pFileQuality: " + pFileQuality)
	log.Printf("PlaylistItems: " + PlaylistItems)
	log.Printf("pYouTubeURL: " + pYouTubeURL)
	log.Printf("pPushoverAppToken: " + pPushoverAppToken)
	log.Printf("pPushoverUserToken: " + pPushoverUserToken)
	log.Println("-----		")

	// =========================================================
	// ============= Download Channel JSON Only ================
	// =========================================================

	// dlname := pChannelID + "/Season_1/" + pChannelID + ".%(ext)s"

	// log.Println("-----		")
	// log.Println("-----		Start Download Channel JSON Only")
	// log.Println("-----		")

	// // out, err := exec.Command("yt-dlp", "-v", "-o", fmt.Sprintf("%s/%s", sMediaFolder, dlname), "--playlist-items", "0", "--write-info-json", "--restrict-filenames", "--add-metadata", "--merge-output-format", pFileFormat, "--format", pFileQuality, "--abort-on-error", "--abort-on-unavailable-fragment", "--no-overwrites", "--continue", pYouTubeURL).Output()

	// out := exec.Command("yt-dlp", "-v", "-o", sMediaFolder+dlname, "--playlist-items", "0", "--write-info-json", "--restrict-filenames", "--add-metadata", "--merge-output-format", pFileFormat, "--format", pFileQuality, "--abort-on-error", "--abort-on-unavailable-fragment", "--no-overwrites", "--continue", pYouTubeURL)
	// out.Stdout = os.Stdout
	// out.Stderr = os.Stderr

	// if err := out.Run(); err != nil {
	// 	log.Printf("------------------      START YT-DLP Channel JSON Only ERROR")
	// 	log.Fatal(err.Error())
	// 	log.Printf("------------------      END YT-DLP Channel JSON Only ERROR")

	// }
	// =========================================================
	// ============= Download Videos with yt-dlp ===============
	// =========================================================

	// dlname2 := pChannelID + "/Season_1/s01e" + channelEpisodeNumberStr + " - %(id)s.%(ext)s"
	dlname2 := pChannelID + "/Season_1/%(id)s.%(ext)s"

	log.Println("-----		")
	log.Println("-----		Download Videos with yt-dlp")
	log.Println("-----		")

	out2 := exec.Command("yt-dlp", "-v", "-o", sMediaFolder+dlname2, "--playlist-items", PlaylistItems, "--write-info-json", "--no-write-playlist-metafiles", "--download-archive", pDownloadArchive, "--restrict-filenames", "--add-metadata", "--merge-output-format", pFileFormat, "--format", pFileQuality, "--abort-on-error", "--abort-on-unavailable-fragment", "--no-overwrites", "--continue", "--write-description", pYouTubeURL)
	out2.Stdout = os.Stdout
	out2.Stderr = os.Stderr

	if err := out2.Run(); err != nil {
		log.Printf("------------------      START YT-DLP ERROR")
		log.Fatal(err.Error())
		log.Printf("------------------      END YT-DLP ERROR")

	}

	// =========================================================
	// ================ List Downloaded Files ==================
	// =========================================================

	log.Println("-----		")
	log.Println("-----		List Downloaded Files")
	log.Println("-----		")
	directory := sMediaFolder + pChannelID
	descfiles, descerr := WalkMatch(directory+"/", "*.description")

	if descerr != nil {
		log.Printf("------------------      START List Downloaded Files ERROR")
		// log.Printf("%s", descerr)
		log.Fatal(descerr)
		log.Printf("------------------      END List Downloaded Files ERROR")
	}

	log.Println("-----		")
	log.Println("-----		List Files to add to RSS Feed")
	log.Println("-----		")
	for _, fname := range descfiles {
		// ------- Get Files ---------
		arrfname_noext := strings.Split(fname, ".")
		fname_noext := arrfname_noext[0]
		fname_json := fname_noext + ".info.json"
		fname_mp3 := fname_noext + ".mp3"
		fname_mp4 := fname_noext + ".mp4"
		fname_description := fname_noext + ".description"

		log.Println("fname_noext: " + fname_noext)
		log.Println("fname_mp3: " + fname_mp3)
		log.Println("fname_mp4: " + fname_mp4)
		log.Println("fname_description: " + fname_description)
		log.Println("fname_json: " + fname_json)

		//  Check if Paths are Valid --
		filename_json_isfile := IsValid(fname_json)
		filename_mp3_isfile := IsValid(fname_mp3)
		filename_mp4_isfile := IsValid(fname_mp4)

		if filename_json_isfile == true {
			log.Println("The JSON file is present.")
		}
		if filename_mp3_isfile == true {
			log.Println("The MP3 file is present.")
			// filename_ext := fname_mp3
		}
		if filename_mp4_isfile == true {
			log.Println("The MP4 file is present.")
			// filename_ext := fname_mp4
		}

		log.Println("-----		")
		log.Println("-----		Get JSON Information")
		log.Println("-----		")

		if filename_json_isfile == true && filename_mp4_isfile == true {
			// //  Open and Read JSON file --
			// Let's first read the `config.json` file
			content, contenterr := ioutil.ReadFile(fname_json)
			if contenterr != nil {
				log.Fatal("Error when opening file: ", contenterr)
			}

			// defining a map
			var mapresult map[string]interface{}
			maperr := json.Unmarshal([]byte(content), &mapresult)

			if maperr != nil {
				// print out if error is not nil
				// fmt.Println(maperr)
				log.Fatal("Error reading JSON File ", maperr)
			}

			var jsonpayload JsonData
			jsonpayload.channel_url = ""
			jsonpayload.description = ""
			jsonpayload.duration_string = "0:0"
			jsonpayload.id = ""
			jsonpayload.thumbnail = ""
			jsonpayload.title = ""
			jsonpayload.uploader_url = ""
			jsonpayload.webpage_url = ""
			// jsonpayload.filesize_approx = 0.0

			jsonpayload.id = fmt.Sprint(mapresult["id"])
			jsonpayload.title = fmt.Sprint(mapresult["title"])
			jsonpayload.thumbnail = fmt.Sprint(mapresult["thumbnail"])
			jsonpayload.description = fmt.Sprint(mapresult["description"])
			jsonpayload.uploader_url = fmt.Sprint(mapresult["uploader_url"])
			jsonpayload.channel_url = fmt.Sprint(mapresult["channel_url"])
			jsonpayload.webpage_url = fmt.Sprint(mapresult["webpage_url"])
			jsonpayload.duration_string = fmt.Sprint(mapresult["duration_string"])
			// jsonpayload.filesize_approx = mapresult["filesize_approx"].(float64)
			// var Filesize float64
			// Filesize = (float64(jsonpayload.filesize_approx) / 1024) / 1024
			// jsonpayload.filesize_approx = roundFloat(Filesize, 2)

			// -- Test Thumbnail Path ----
			ytvideo_thumbnail := "https://i.ytimg.com/vi_webp/" + jsonpayload.id + "/maxresdefault.webp"
			ValidURI := IsValidURL(ytvideo_thumbnail)
			if ValidURI == true {
				jsonpayload.thumbnail = ytvideo_thumbnail
			}

			ytvideo_thumbnail2 := "https://i.ytimg.com/vi_webp/" + jsonpayload.id + "/maxresdefault.jpg"
			ValidURI2 := IsValidURL(ytvideo_thumbnail2)
			if ValidURI2 == true {
				jsonpayload.thumbnail = ytvideo_thumbnail2
			}

			// =========================================================
			// ============ Validate Episode Number File ===============
			// =========================================================

			channelEpisodeNumberPath := Config + pChannelID + "_EpisodeNumber.txt"
			channelEpisodeNumberPath_Valid := IsValid(channelEpisodeNumberPath)

			if channelEpisodeNumberPath_Valid == false {
				if writersserr := os.WriteFile(channelEpisodeNumberPath, []byte("0"), 0666); writersserr != nil {
					log.Fatal(writersserr)
				}
			}

			epContent, epErr := ioutil.ReadFile(channelEpisodeNumberPath) // the file is inside the local directory
			if epErr != nil {
				log.Fatal(epErr)
			}
			channelEpisodeNumbertmp := strings.TrimSpace(string(epContent))
			channelEpisodeNumber, interr := strconv.ParseInt(channelEpisodeNumbertmp, 10, 64)
			channelEpisodeNumber += 1
			channelEpisodeNumberStr := fmt.Sprintf("%02d", channelEpisodeNumber)

			if interr != nil {
				log.Printf("------------------      START strconv.ParseInt ERROR")
				log.Fatal(interr.Error())
				log.Printf("------------------      END strconv.ParseInt ERROR")
			}

			log.Printf("channelEpisodeNumberPath: " + channelEpisodeNumberPath)
			log.Printf("channelEpisodeNumber_Valid: " + fmt.Sprint(channelEpisodeNumberPath_Valid))
			log.Printf("channelEpisodeNumber: '" + fmt.Sprint(channelEpisodeNumber) + "'")
			log.Printf("channelEpisodeNumberStr: '" + channelEpisodeNumberStr + "'")

			// ~~~~~~~ Write New Episode Number ~~~~~~~~~
			if writersserr := os.WriteFile(channelEpisodeNumberPath, []byte(fmt.Sprint(channelEpisodeNumber)), 0666); writersserr != nil {
				log.Fatal(writersserr)
			}

			// ~~~~~~ Download Episode Thumbnail ~~~~~~~~

			savename := ""
			if strings.HasSuffix(jsonpayload.thumbnail, ".jpg") {
				savename = "s01e" + channelEpisodeNumberStr + " - " + jsonpayload.id + ".jpg"
			}

			if strings.HasSuffix(jsonpayload.thumbnail, ".webp") {
				savename = "s01e" + channelEpisodeNumberStr + " - " + jsonpayload.id + ".webp"
			}

			if strings.HasSuffix(jsonpayload.thumbnail, ".jpg") == false && strings.HasSuffix(jsonpayload.thumbnail, ".webp") == false {
				savename = "s01e" + channelEpisodeNumberStr + " - " + jsonpayload.id + ".jpg"
			}

			err := DownloadFile(sMediaFolder+pChannelID+"/Season_1/"+savename, jsonpayload.thumbnail)
			if err != nil {
				// panic(err)
				log.Printf("------------------      START DownloadFile ERROR")
				log.Fatal(err.Error())
				log.Printf("------------------      END DownloadFile ERROR")
			}
			fmt.Println("Downloaded: " + jsonpayload.thumbnail)

			// ~~~~~~~~~~~ Rename MP4 File ~~~~~~~~~~~~~~

			// dlname2 := pChannelID + "/Season_1/s01e" + channelEpisodeNumberStr + " - %(id)s.%(ext)s"
			os.Rename(sMediaFolder+pChannelID+"/Season_1/"+jsonpayload.id+".mp4", sMediaFolder+pChannelID+"/Season_1/s01e"+channelEpisodeNumberStr+" - "+jsonpayload.id+".mp4")

			// --- Print Final Data ------

			log.Printf("jsonpayload.id: " + jsonpayload.id)
			log.Printf("jsonpayload.title: " + jsonpayload.title)
			log.Printf("jsonpayload.thumbnail: " + jsonpayload.thumbnail)
			// log.Printf("jsonpayload.description: " + jsonpayload.description)
			log.Printf("jsonpayload.uploader_url: " + jsonpayload.uploader_url)
			log.Printf("jsonpayload.channel_url: " + jsonpayload.channel_url)
			log.Printf("jsonpayload.webpage_url: " + jsonpayload.webpage_url)
			log.Printf("jsonpayload.duration_string: " + jsonpayload.duration_string)
			// log.Printf("jsonpayload.filesize_approx: " + fmt.Sprint(jsonpayload.filesize_approx))

			// =========================================================
			// =================== Notify Pushover =====================
			// =========================================================

			NotifyPushover(Config, pPushoverAppToken, pPushoverUserToken, "RSS Podcast Downloaded ("+pName+")", "<html><body>"+jsonpayload.title+"<br /><br />--------------------------------------------<br /><br />"+jsonpayload.description+"</body></html>", jsonpayload.thumbnail, jsonpayload.webpage_url)
		}
	}
}

func main() {
	// name := "Go Developers"
	// log.Println("Hello World:", name)
	xmlFile, err := os.Open("/config/settings.xml")
	// xmlFile, err := os.Open("settingsLOCAL.xml")
	if err != nil {
		log.Println(err)
	}
	// log.Println("Successfully Opened users.xml")

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(xmlFile)

	// we initialize our PodcastDownload array
	var settingsXML settings
	var validateXML Validate
	// we unmarshal our byteArray which contains our
	// xmlFiles content into 'users' which we defined above
	xml.Unmarshal(byteValue, &settingsXML)

	log.Println("Email: " + settingsXML.Email)
	log.Println("MediaFolder: " + settingsXML.MediaFolder)
	log.Println("PushoverUserToken: " + settingsXML.PushoverUserToken)
	log.Println("Config: " + settingsXML.Config)

	// =========================================================
	// ================== Validate Settings ====================
	// =========================================================

	log.Println()
	validateXML.MediaFolder = IsValid(settingsXML.MediaFolder)
	validateXML.Config = IsValid(settingsXML.Config)

	if settingsXML.PlaylistItems == "" {
		validateXML.PlaylistItems = false
	} else {
		validateXML.PlaylistItems = true
	}

	if settingsXML.PushoverUserToken == "" {
		validateXML.PushoverUserToken = false
	} else {
		validateXML.PushoverUserToken = true
	}

	// ~~~~~~~~ Print Validation Data ~~~~~~~~~~~

	log.Println("MediaFolder Valid: " + fmt.Sprint(validateXML.MediaFolder))
	log.Println("Config Valid: " + fmt.Sprint(validateXML.Config))
	log.Println("PushoverUserToken Valid: " + fmt.Sprint(validateXML.PushoverUserToken))
	log.Println("PlaylistItems Valid: " + fmt.Sprint(validateXML.PlaylistItems))

	// =========================================================
	// =========================================================
	// =========================================================

	// ########################################################################
	// ######################## Loop PodcastDownload ##########################
	// ########################################################################

	if validateXML.MediaFolder == true && validateXML.Config == true && validateXML.PushoverUserToken == true && validateXML.PlaylistItems == true {
		log.Println("-----		")
		log.Println("-----		Start Validate")
		log.Println("-----		")
		log.Println("Valid - MediaFolder")
		log.Println("Valid - RSSFolder")
		log.Println("Valid - RSSTemplate")
		log.Println("Valid - Config")
		// we iterate through every user within our users array and
		// print out the user Type, their name, and their facebook url
		// as just an example
		for i := 0; i < len(settingsXML.PodcastDownload); i++ {
			if settingsXML.PodcastDownload[i].Name == "" && settingsXML.PodcastDownload[i].ChannelID == "" && settingsXML.PodcastDownload[i].ChannelThumbnail == "" && settingsXML.PodcastDownload[i].DownloadArchive == "" && settingsXML.PodcastDownload[i].FileFormat == "" && settingsXML.PodcastDownload[i].FileQuality == "" && settingsXML.PlaylistItems == "" && settingsXML.PodcastDownload[i].YouTubeURL == "" && settingsXML.PodcastDownload[i].PushoverAppToken == "" {
				validateXML.PodcastDownload_Name = false
				validateXML.PodcastDownload_ChannelID = false
				validateXML.PodcastDownload_DownloadArchive = false
				validateXML.PodcastDownload_FileFormat = false
				validateXML.PodcastDownload_FileQuality = false
				validateXML.PodcastDownload_YouTubeURL = false
				validateXML.PlaylistItems = false
				log.Println("Not Valid - PodcastDownload_Name")
				log.Println("Not Valid - PodcastDownload_ChannelID")
				log.Println("Not Valid - PodcastDownload_DownloadArchive")
				log.Println("ot Valid - PodcastDownload_FileFormat")
				log.Println("Not Valid - PodcastDownload_FileQuality")
				log.Println("Not Valid - PodcastDownload_YouTubeURL")
				log.Println("Not Valid - PlaylistItems")
			} else {
				validateXML.PodcastDownload_DownloadArchive = IsValid(settingsXML.PodcastDownload[i].DownloadArchive)
				if validateXML.PodcastDownload_DownloadArchive == true {
					validateXML.PodcastDownload_Name = true
					validateXML.PodcastDownload_ChannelID = true
					validateXML.PodcastDownload_FileFormat = true
					validateXML.PodcastDownload_FileQuality = true
					validateXML.PodcastDownload_YouTubeURL = true
					validateXML.PlaylistItems = true
					log.Println("Valid - PodcastDownload_Name")
					log.Println("Valid - PodcastDownload_ChannelID")
					log.Println("Valid - PodcastDownload_DownloadArchive")
					log.Println("Valid - PodcastDownload_FileFormat")
					log.Println("Valid - PodcastDownload_FileQuality")
					log.Println("Valid - PodcastDownload_YouTubeURL")
					log.Println("Valid - PlaylistItems")
				} else {
					log.Println("Valid - PodcastDownload_Name")
					log.Println("Valid - PodcastDownload_ChannelID")
					log.Println("Not Valid - PodcastDownload_DownloadArchive")
					log.Println("Valid - PodcastDownload_FileFormat")
					log.Println("Valid - PodcastDownload_FileQuality")
					log.Println("Valid - PodcastDownload_YouTubeURL")
					log.Println("Valid - PlaylistItems")
				}
			}
			log.Println("-----		")
			log.Println("-----		End Validate")
			log.Println("-----		")
			log.Println("")

			// =========================================================
			// =================== Check All Valid =====================
			// =========================================================

			if validateXML.MediaFolder == true && validateXML.PodcastDownload_ChannelID == true && validateXML.PodcastDownload_DownloadArchive == true && validateXML.PodcastDownload_FileFormat == true && validateXML.PodcastDownload_FileQuality == true && validateXML.PodcastDownload_Name == true && validateXML.PlaylistItems == true && validateXML.PodcastDownload_YouTubeURL == true {
				log.Println("-----		")
				log.Println("-----		PodcastDownload")
				log.Println("-----		")
				log.Println("PodcastDownload.Name: " + settingsXML.PodcastDownload[i].Name)
				log.Println("PodcastDownload.ChannelID: " + settingsXML.PodcastDownload[i].ChannelID)
				log.Println("PodcastDownload.ChannelThumbnail: " + settingsXML.PodcastDownload[i].ChannelThumbnail)
				log.Println("PodcastDownload.DownloadArchive: " + settingsXML.PodcastDownload[i].DownloadArchive)
				log.Println("PodcastDownload.FileFormat: " + settingsXML.PodcastDownload[i].FileFormat)
				log.Println("PodcastDownload.FileQuality: " + settingsXML.PodcastDownload[i].FileQuality)
				log.Println("PodcastDownload.YouTubeURL: " + settingsXML.PodcastDownload[i].YouTubeURL)
				log.Println("PlaylistItems: " + settingsXML.PlaylistItems)
				log.Println("-----		")

				Run_YTDLP(settingsXML.MediaFolder, settingsXML.Config, settingsXML.PodcastDownload[i].Name, settingsXML.PodcastDownload[i].ChannelID, settingsXML.PodcastDownload[i].FileFormat, settingsXML.PodcastDownload[i].DownloadArchive, settingsXML.PodcastDownload[i].FileQuality, settingsXML.PlaylistItems, settingsXML.PodcastDownload[i].YouTubeURL, settingsXML.PodcastDownload[i].PushoverAppToken, settingsXML.PushoverUserToken)
				DeleteOldFiles(settingsXML.MediaFolder + settingsXML.PodcastDownload[i].ChannelID + "/")
				log.Println("")
			}
		}
	}

	// ########################################################################
	// ########################################################################
	// ########################################################################

	// defer the closing of our xmlFile so that we can parse it later on
	defer xmlFile.Close()
}
