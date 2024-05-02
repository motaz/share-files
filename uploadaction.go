package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/motaz/codeutils"
)

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
		subdir := suggestDirctory(userkey)
		shareroot := codeutils.GetConfigValue("config.ini", "shareroot")
		if shareroot == "" {
			shareroot = getHomeDirectory() + "/share/"
		}
		sharedir := shareroot + subdir
		afilename := sharedir + "/" + handler.Filename
		checkSubdir(sharedir)
		f, err := os.OpenFile(afilename, os.O_WRONLY|os.O_CREATE, 0666)
		defer f.Close()
		expire := time.Now()

		// Expiary
		days := r.FormValue("days")
		daysInt, _ := strconv.Atoi(days)
		if daysInt > 90 {
			daysInt = 90

			expire = expire.AddDate(0, 0, daysInt)
		}

		infoFilename := afilename + ".expire"

		fileinfo := fileInfoType{Filename: afilename, Expires: expire, UserKey: userkey,
			ShareKey: sharekey, Filenameonly: handler.Filename, Subdir: subdir}
		jsonData, err := json.Marshal(fileinfo)

		err = ioutil.WriteFile(infoFilename, jsonData, 0644)
		if err != nil {
			w.Write([]byte("<h3>Error: " + err.Error() + "</h3>"))

		} else {
			io.Copy(f, file)

			link := getCommonShareLink(r) + subdir + "/" + handler.Filename
			fileinfo.Link = link

			mytemplate.ExecuteTemplate(w, "result.html", fileinfo)

			writeLog("Received : " + handler.Filename + ", from " + r.RemoteAddr)

		}
	}
}

func checkSubdir(subdir string) {

	if !codeutils.IsFileExists(subdir) {
		os.MkdirAll(subdir, os.ModePerm)
		ioutil.WriteFile(subdir+"/index.html", nil, os.ModePerm)

	}
}
