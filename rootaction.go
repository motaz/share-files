package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type FileInfoType struct {
	Entry        string
	Filename     string
	Filenameonly string
	Expires      time.Time
	Link         string
	UserKey      string
	ShareKey     string
	Subdir       string
	Size         int64
}

type uploadForm struct {
	Key string
}

func removeFiles(userkey string, w http.ResponseWriter, req *http.Request) {

	list := req.Form["fileid"]
	for _, file := range list {
		info := readMetaFile(file + ".expire")
		if info.UserKey == userkey {

			err := os.Remove(file)
			if err == nil || strings.Contains(err.Error(), "no such") {
				err = os.Remove(file + ".expire")
			}
			if err != nil {
				fmt.Fprintln(w, "<font color=red>Error deleting file: "+file+" : "+
					err.Error()+"</font></br>")
			} else {

				fmt.Fprintln(w, "File : "+file+" has been removed</br>")
			}
		}
	}

}

func viewUpload(w http.ResponseWriter, req *http.Request) {

	checkIndexFile(req)
	searchkey := req.FormValue("searchkey")
	var uform uploadForm
	uform.Key = searchkey

	err := mytemplate.ExecuteTemplate(w, "index.html", uform)

	if err != nil {
		w.Write([]byte("Error: " + err.Error()))
	}

	userkey := getUserKey(w, req)

	if req.FormValue("remove") != "" {
		removeFiles(userkey, w, req)
	}
	w.Write([]byte("<div  class='well'>"))

	var files []FileInfoType
	if searchkey != "" {
		w.Write([]byte("<h3>Shared key Files</h3>"))

		files = listFiles("", searchkey)
		displayFiles(files, w, req, false)

	} else {
		w.Write([]byte("<h3>Your Previous Files</h3>"))

		files = listFiles(userkey, "")

		displayFiles(files, w, req, true)

	}

}

type FileType struct {
	Fileid    string
	Filename  string
	Filetime  string
	Filelink  string
	Sharekey  string
	Candelete bool
}

type DisplayData struct {
	Files     []FileType
	Candelete bool
	Host      string
}

func displayFiles(files []FileInfoType, w http.ResponseWriter, r *http.Request, canDelete bool) {

	fmt.Fprintln(w, "<form method=post>")
	fmt.Fprintln(w, "<table>")

	share := getCommonShareLink(r)
	var list []FileType
	var displayData DisplayData
	displayData.Candelete = canDelete
	displayData.Host = r.Host

	for _, file := range files {
		var afile FileType
		afile.Candelete = canDelete
		afile.Fileid = file.Filename
		afile.Sharekey = file.ShareKey

		subdir := ""
		if file.Subdir != "" {
			subdir = file.Subdir + "/"
		}
		afile.Filelink = share + subdir + file.Filenameonly
		afile.Filename = file.Filenameonly

		afile.Filetime = file.Expires.String()[:19]

		list = append(list, afile)
	}

	displayData.Files = list
	err := mytemplate.ExecuteTemplate(w, "files.html", displayData)
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

}
