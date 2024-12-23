package main

import (
	"bufio"
	"bytes"
	"io"

	"net/http"

	"strconv"
	"time"

	"github.com/motaz/codeutils"
	"github.com/motaz/redisaccess"
)

const FILE_INFO_KEY = "share-files::info::"
const FILE_CONTENT_KEY = "share-files::file::"

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
	redisaccess.SetBytes(FILE_CONTENT_KEY+entry, buf.Bytes(), dur)
	fileinfo := FileInfoType{Entry: entry, Filename: handler.Filename,
		Expires: expire, UserKey: userkey,
		ShareKey: sharekey, Filenameonly: handler.Filename, Size: size, Uploadtime: time.Now()}
	if sharekey != "" {
		fileinfo.ShareKeyHash = codeutils.GetMD5(sharekey + userkey + "10022-0#2")
	}
	redisaccess.SetValue(FILE_INFO_KEY+entry, fileinfo, dur)

	mytemplate.ExecuteTemplate(w, "result.html", fileinfo)

	writeLog("Received : " + handler.Filename + ", from " + r.RemoteAddr)

}
