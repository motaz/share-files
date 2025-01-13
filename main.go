// share-files project main.go
// Started at 18.Feb.2018 by Motaz Abdel Azeem
// Updated on 13.Jan.2025

package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"

	"github.com/motaz/redisaccess"
)

var mytemplate *template.Template

//go:embed templates
var templates embed.FS

//go:embed resources
var static embed.FS

func InitTemplate(embededTemplates embed.FS) error {
	var err error
	mytemplate, err = template.ParseFS(embededTemplates, "templates/*.html")
	if err != nil {
		fmt.Println("error in InitTemplate: " + err.Error())
		return err
	}
	return nil
}

func main() {
	InitTemplate(templates)
	_, err := redisaccess.InitRedisLocalhost()
	if err != nil {
		fmt.Println("Redis error: ", err.Error())
	} else {

		http.HandleFunc("/", redirectToIndex)
		http.HandleFunc("/share-files", viewUpload)
		http.HandleFunc("/share-files/", viewUpload)
		http.HandleFunc("/share-files/up", upload)
		http.HandleFunc("/share-files/view", viewFile)
		http.HandleFunc("/share-files/fileinfo", viewFileInfo)

		http.Handle("/share-files/resources/", http.StripPrefix("/share-files/", http.FileServer(http.FS(static))))

		fmt.Println("ReceiveFile, Listening on port 10026")
		fmt.Println("http://localhost:10026")
		err := http.ListenAndServe(":10026", nil)
		if err != nil {
			fmt.Println("Error while listening: ", err.Error())
		}
	}
}

func redirectToIndex(w http.ResponseWriter, req *http.Request) {

	http.Redirect(w, req, "/share-files"+req.RequestURI, http.StatusTemporaryRedirect)
}
