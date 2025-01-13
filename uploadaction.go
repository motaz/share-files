package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"net/http"

	"strconv"
	"time"

	"github.com/motaz/codeutils"
	"github.com/motaz/redisaccess"
)

const FILE_INFO_KEY = "share-files::info::"
const FILE_CONTENT_KEY = "share-files::file::"
const FILE_CLIENT_KEY = "share-files::client::-"

type FileCountType struct {
	Count int
	Size  int
}

func incrementCount(IP string, size int) {

	var fileCount FileCountType
	redisaccess.ReadValue(FILE_CLIENT_KEY+IP, &fileCount)
	fileCount.Count++
	fileCount.Size += size

	redisaccess.SetValue(FILE_CLIENT_KEY+IP, fileCount, time.Hour)

}

func isAllowed(w http.ResponseWriter, clientIP string) (allow bool) {

	allow = true

	// Check upload limit for this IP
	var fileCount FileCountType
	found, err := redisaccess.ReadValue(FILE_CLIENT_KEY+clientIP, &fileCount)
	if !found {
		fileCount.Count = 0
		fileCount.Size = 0
	}
	fileCount.Count++

	if err == nil {

		limit, _ := codeutils.ReadINIAsInt("config.ini", "", "limitperhour")
		if limit < 1 {
			limit = 20
		}
		size, _ := codeutils.ReadINIAsInt("config.ini", "", "sizelimit")
		if size < 1 {
			size = 512_000_000
		}
		fmt.Println(size, limit)

		if fileCount.Count >= limit || fileCount.Size > size {
			writeLog(fmt.Sprintf("Client %s has exceeded limit: %+v ", clientIP, fileCount))
			http.Error(w, "Too many files or upload size limit per hour reached", http.StatusTooManyRequests)
			allow = false
		}
	}
	return

}

// upload logic
func upload(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("Content-Type", "text/html;charset=UTF-8")
	w.Header().Add("encoding", "UTF-8")

	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("uploadfile")
	if err != nil {
		w.Write([]byte("<h3>Error: " + err.Error() + "</h3>"))
		return
	}
	defer file.Close()
	userkey := getUserKey(w, r)
	sharekey := r.FormValue("key")
	clientIP := codeutils.GetRemoteIP(r)

	if isAllowed(w, clientIP) {
		toHash := sharekey
		if toHash == "" {
			toHash = "default"
		}
		entry := suggestEntryID(userkey)

		expire := time.Now()

		// Expiary
		days := r.FormValue("days")
		daysInt, _ := strconv.Atoi(days)
		maxLimit := readKeeplimit()

		if daysInt > maxLimit {
			daysInt = maxLimit
		}

		expire = expire.AddDate(0, 0, daysInt)

		var dur time.Duration
		atime := int(time.Hour) * daysInt * 24
		dur = time.Duration(atime)

		filer := bufio.NewReader(file)
		buf := bytes.NewBuffer(nil)
		size, err := io.Copy(buf, filer)
		if err == nil {
			redisaccess.SetBytes(FILE_CONTENT_KEY+entry, buf.Bytes(), dur)
			fileinfo := FileInfoType{Entry: entry, Filename: handler.Filename,
				Expires: expire, UserKey: userkey,
				ShareKey: sharekey, Filenameonly: handler.Filename, Size: size, Uploadtime: time.Now(),
				IP: clientIP}
			if sharekey != "" {
				spice := readIniValue("", "spice", "10022-0#2")
				fileinfo.ShareKeyHash = codeutils.GetMD5(sharekey + userkey + spice)
			}
			redisaccess.SetValue(FILE_INFO_KEY+entry, fileinfo, dur)
			incrementCount(clientIP, int(size))

			mytemplate.ExecuteTemplate(w, "result.html", fileinfo)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		writeLog("Received : " + handler.Filename + ", from " + clientIP)
	}

}
