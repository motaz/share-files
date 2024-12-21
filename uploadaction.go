package main

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"os"
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
	if r.Method == "GET" {
		//crutime := time.Now().Unix()
		//h := md5.New()
		//o.WriteString(h, strconv.FormatInt(crutime, 10))

		//	t, _ := template.ParseFiles("upload.gtpl")
		//t.Execute(w, token)
	} else {
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
		entry := suggestDirctory(userkey)

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
		redisaccess.SetValue(FILE_CONTENT_KEY+entry, buf.Bytes(), dur)
		fileinfo := FileInfoType{Entry: entry, Filename: handler.Filename,
			Expires: expire, UserKey: userkey,
			ShareKey: sharekey, Filenameonly: handler.Filename, Size: size}

		redisaccess.SetValue(FILE_INFO_KEY+entry, fileinfo, dur)

		//link := getCommonShareLink(r) + subdir + "/" + handler.Filename
		//	fileinfo.Link

		mytemplate.ExecuteTemplate(w, "result.html", fileinfo)

		writeLog("Received : " + handler.Filename + ", from " + r.RemoteAddr)

	}
}

func checkSubdir(subdir string) {

	if !codeutils.IsFileExists(subdir) {
		os.MkdirAll(subdir, os.ModePerm)
		ioutil.WriteFile(subdir+"/index.html", nil, os.ModePerm)

	}
}
