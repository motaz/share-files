package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/motaz/codeutils"
)

func getRandom(r int) int {

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return r1.Intn(r)
}

func getUserKey(w http.ResponseWriter, req *http.Request) (userKey string) {

	userCookie, err := req.Cookie("userkey")
	secondCookie, err2 := req.Cookie("secondkey")
	var secondKey string

	if err == nil {
		userKey = userCookie.Value
		secondKey = userKey
	} else if userKey == "" {
		if err2 == nil {
			userKey = secondCookie.Value
		} else {
			userKey = setCookies(w, req)
		}
	}
	if err2 != nil && err == nil {
		secondKey = userKey
		expiration2 := time.Now().Add(time.Hour * 24 * 100)
		cookie2 := http.Cookie{Name: "secondkey", Value: secondKey, Expires: expiration2}

		http.SetCookie(w, &cookie2)

	}
	return
}

func setCookies(w http.ResponseWriter, r *http.Request) (userKey string) {

	expiration := time.Now().Add(time.Hour * 24 * 90)
	expiration2 := time.Now().Add(time.Hour * 24 * 100)
	userKey = strconv.Itoa(getRandom(10000000))
	cookie := http.Cookie{Name: "userkey", Value: userKey, Expires: expiration}
	cookie2 := http.Cookie{Name: "secondkey", Value: userKey, Expires: expiration2}

	http.SetCookie(w, &cookie)
	http.SetCookie(w, &cookie2)
	return
}

func getCommonShareLink(r *http.Request) (share string) {

	serverip := r.Host
	if strings.Contains(serverip, ":") {
		serverip = serverip[:strings.Index(serverip, ":")]
	}

	share = serverip + "/share/"
	return
}

func writeLog(event string) {
	codeutils.WriteToLog(event, "uploaded")
}

func loopCheck() {

	for {
		checkExpiary()
		time.Sleep(time.Hour)

	}
}

func checkExpiary() {

	dirName := getHomeDirectory() + "/share/"
	directories, _ := ioutil.ReadDir(dirName)
	for _, dir := range directories {
		if dir.IsDir() && dir.Name() != "." && dir.Name() != ".." {
			checkDirectoryExpiary(dir.Name())
		}
	}

}

func checkDirectoryExpiary(dir string) {

	searchFiles := getHomeDirectory() + "/share/" + dir + "/*.expire"
	now := time.Now()

	files2, _ := filepath.Glob(searchFiles)
	for _, f2 := range files2 {
		contents, _ := ioutil.ReadFile(f2)
		var info fileInfoType
		json.Unmarshal(contents, &info)

		if now.After(info.Expires) {
			err := os.Remove(info.Filename)
			if err != nil {
				println("Error: " + err.Error())
			}
			os.Remove(f2)
		}

	}
}

func listFiles(userkey string, sharekey string) (files []fileInfoType) {

	searchFiles := codeutils.GetConfigValue("config.ini", "shareroot")
	if searchFiles == "" {

		searchFiles = getHomeDirectory() + "/share/"
	}
	directories, _ := ioutil.ReadDir(searchFiles)
	searchDirectory(searchFiles+"/*.expire", userkey, sharekey, &files)
	for _, dir := range directories {
		if dir.IsDir() {
			searchDirectory(searchFiles+dir.Name()+"/*.expire", userkey, sharekey, &files)

		}
	}

	return
}

func readMetaFile(filename string) fileInfoType {
	contents, _ := ioutil.ReadFile(filename)
	var info fileInfoType
	json.Unmarshal(contents, &info)
	return info

}
func searchDirectory(path, userkey string, sharekey string, files *[]fileInfoType) {

	files2, _ := filepath.Glob(path)
	for _, f2 := range files2 {
		info := readMetaFile(f2)

		if ((userkey != "" && info.UserKey == userkey) ||
			(sharekey != "" && info.ShareKey == sharekey)) &&
			info.Filenameonly != "" {
			*files = append(*files, info)

		}
	}

}

func getHomeDirectory() string {

	currentdir, _ := user.Current()

	return currentdir.HomeDir
}

func suggestDirctory(sharekey string) string {

	hash := getMD5Hash(sharekey)
	suggestDir := hash[:5]
	return suggestDir
}

func getMD5Hash(text string) string {

	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func checkIndexFile(r *http.Request) {
	location := r.Referer()
	if location == "" {
		if r.URL.Port() != "80" && r.URL.Port() != "443" {
			port := r.URL.Port()
			port = strings.ReplaceAll(port, ":", "")
			location = "http://" + r.Host + "/upload"
			println(location)
			println("host ", r.Host)
		} else {
			location = "/upload"
		}

	} else {
		if !strings.HasSuffix(location, "/upload") {
			location += "/upload"
		}
	}
	contents := `<html>
<head>
<meta http-equiv="Cache-Control" content="no-cache, no-store, must-revalidate" />
<meta http-equiv="Pragma" content="no-cache" />
<meta http-equiv="Expires" content="0" />
</head>
<body>
<script>location="` + location + `"
</script>
</body>
</html>
`
	shareroot := codeutils.GetConfigValue("config.ini", "shareroot")
	indexFile := shareroot + "index.html"
	if !codeutils.IsFileExists(indexFile) {

		os.WriteFile(indexFile, []byte(contents), os.ModePerm)

	}
}
