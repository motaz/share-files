package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net"

	"net/http"

	"strconv"
	"time"

	"github.com/motaz/codeutils"
	"github.com/motaz/redisaccess"
)

const FILE_INFO_KEY = "share-files::info::"
const FILE_CONTENT_KEY = "share-files::file::"
const FILE_CLIENT_KEY = "share-files::client::"

type FileCountType struct {
	Count int
	Size  int64
}

func incrementCount(IP string, size int64) {

	data, found, err := redisaccess.GetBytes(FILE_CLIENT_KEY + clientIP)
	var fileCount FileCountType
	if found {
		json.Unmarshal(data, fileCount)
	}
	fileCount.Count++
	fileCount.Size += size
	content, _ := json.Marshal(fileCount)
	redisaccess.SetBytes(FILE_CLIENT_KEY+clientIP, content)

}

func isAllowed(w http.ResponseWriter, r *http.Request) (allow bool) {

	allow = true
	clientIP, _, _ := net.SplitHostPort(r.RemoteAddr)

	// Check upload limit for this IP
	uploadCountStr, found, err := redisaccess.GetValue(FILE_CLIENT_KEY + clientIP)
	if !found {
		uploadCountStr = "0"
	}
	uploadCount, err := strconv.Atoi(uploadCountStr)
	if err == nil {

		limit, _ := codeutils.ReadINIAsInt("config.ini", "", "limitperhour")
		if limit < 1 {
			limit = 10
		}

		if uploadCount >= limit {
			http.Error(w, "Too many uploads from this IP in the last hour", http.StatusTooManyRequests)
			allow = false
		} else {
			uploadCount++
			redisaccess.SetValue(FILE_CLIENT_KEY+clientIP, uploadCount, time.Hour)
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
	if isAllowed(w, r) {
		toHash := sharekey
		if toHash == "" {
			toHash = "default"
		}
		entry := suggestEntryID(userkey)

		expire := time.Now()

		// Expiary
		days := r.FormValue("days")
		daysInt, _ := strconv.Atoi(days)
		if daysInt > 90 {
			daysInt = 90
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
			ip := codeutils.GetRemoteIP(r)
			fileinfo := FileInfoType{Entry: entry, Filename: handler.Filename,
				Expires: expire, UserKey: userkey,
				ShareKey: sharekey, Filenameonly: handler.Filename, Size: size, Uploadtime: time.Now(),
				IP: ip}
			if sharekey != "" {
				fileinfo.ShareKeyHash = codeutils.GetMD5(sharekey + userkey + "10022-0#2")
			}
			redisaccess.SetValue(FILE_INFO_KEY+entry, fileinfo, dur)

			mytemplate.ExecuteTemplate(w, "result.html", fileinfo)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		writeLog("Received : " + handler.Filename + ", from " + r.RemoteAddr)
	}

}
