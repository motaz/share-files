package main

import (
	"encoding/json"
	"fmt"
	"sort"

	"net/http"

	"time"

	"github.com/motaz/redisaccess"
)

type FileInfoType struct {
	Entry        string
	Filename     string
	Filenameonly string
	Uploadtime   time.Time
	Expires      time.Time
	Link         string
	UserKey      string
	ShareKey     string
	ShareKeyHash string
	Subdir       string
	Size         int64
	Downloads    int
}

type uploadForm struct {
	Key string
}

func removeFiles(userkey string, w http.ResponseWriter, req *http.Request) (success bool) {

	list := req.Form["fileid"]
	success = true
	for _, entry := range list {
		infoBytes, found, _ := redisaccess.GetBytes(FILE_INFO_KEY + entry)
		if found {
			var info FileInfoType
			json.Unmarshal(infoBytes, &info)
			if info.UserKey == userkey {

				err := redisaccess.RemoveValue(FILE_INFO_KEY + entry)
				redisaccess.RemoveValue(FILE_CONTENT_KEY + entry)
				if err != nil {
					fmt.Fprintln(w, "<font color=red>Error deleting file: "+info.Filename+" : "+
						err.Error()+"</font></br>")
				} else {

					fmt.Fprintln(w, "File : "+info.Filename+" has been removed</br>")
				}
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
			success = false
			fmt.Fprintln(w, "<font color=red>File not found </font></br>")
		}
	}
	return

}

func viewUpload(w http.ResponseWriter, req *http.Request) {

	searchkey := req.FormValue("searchkey")
	sharekey := req.FormValue("sharekey")
	var uform uploadForm
	uform.Key = searchkey
	userkey := getUserKey(w, req)

	var files []FileInfoType
	if searchkey != "" || sharekey != "" {

		searchDirectory("", searchkey, sharekey, &files)
		for _, info := range files {
			if info.UserKey == userkey {
				uform.Key = info.ShareKey
				break
			}

		}

	}
	err := mytemplate.ExecuteTemplate(w, "index.html", uform)

	if err != nil {
		w.Write([]byte("Error: " + err.Error()))
	}

	if req.FormValue("remove") != "" {
		success := removeFiles(userkey, w, req)
		if !success {
			return
		}
	}
	w.Write([]byte("<div  class='well'>"))
	if searchkey != "" || sharekey != "" {
		w.Write([]byte("<h3>Shared key Files</h3>"))
	} else {
		w.Write([]byte("<h3>Your Previous Files</h3>"))

		searchDirectory(userkey, "", "", &files)
	}

	displayFiles(files, w, req, true)

}

type FileType struct {
	Fileid       string
	Filename     string
	Expiretime   string
	Uploadtime   string
	Downloads    int
	Filelink     string
	Sharekey     string
	Sharekeyhash string
	Candelete    bool
}

type DisplayData struct {
	Files     []FileType
	Candelete bool
	Host      string
}

func displayFiles(files []FileInfoType, w http.ResponseWriter, r *http.Request, canDelete bool) {

	fmt.Fprintln(w, "<form method=post>")
	fmt.Fprintln(w, "<table>")

	if len(files) > 0 {
		sort.Slice(files, func(i, j int) bool {
			return files[i].Uploadtime.After(files[j].Uploadtime)
		})
	}
	var list []FileType
	var displayData DisplayData
	displayData.Candelete = canDelete
	displayData.Host = r.Host

	for _, file := range files {
		var afile FileType
		afile.Candelete = canDelete
		afile.Fileid = file.Entry
		afile.Sharekey = file.ShareKey
		afile.Sharekeyhash = file.ShareKeyHash

		afile.Filelink = "view?id=" + file.Entry
		afile.Filename = file.Filenameonly
		afile.Uploadtime = file.Uploadtime.String()[:19]
		afile.Expiretime = file.Expires.String()[:19]
		afile.Downloads = file.Downloads

		list = append(list, afile)
	}

	displayData.Files = list
	err := mytemplate.ExecuteTemplate(w, "files.html", displayData)
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

}

func viewFile(w http.ResponseWriter, req *http.Request) {

	entry := req.FormValue("id")

	fileInfoBytes, found, err := redisaccess.GetBytes(FILE_INFO_KEY + entry)
	if err != nil || !found {

		fmt.Println("Error in viewFile: ", err.Error())
		if !found {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "File not found")
			return

		}
	} else if found {
		var fileInfo FileInfoType
		json.Unmarshal(fileInfoBytes, &fileInfo)
		fileContents, _, err := redisaccess.GetBytes(FILE_CONTENT_KEY + entry)

		if err != nil {
			w.Write([]byte("Error: " + err.Error()))
		} else {
			w.Header().Set("Content-Disposition", "filename="+fileInfo.Filename+";")
			w.Write(fileContents)
			fileInfo.Downloads++
			ttl, err := redisaccess.GetTTL(FILE_INFO_KEY + entry)
			if err == nil {
				redisaccess.SetValue(FILE_INFO_KEY+entry, fileInfo, ttl)
			}

			if err != nil {
				w.Write([]byte("Error copying file: " + err.Error()))
			}
		}
	}

}
