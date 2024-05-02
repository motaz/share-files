// share-files project main.go
// Started at 18.Feb.2018 by Motaz Abdel Azeem

package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
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

	go loopCheck()
	http.HandleFunc("/", redirectToIndex)
	http.HandleFunc("/upload", viewUpload)
	http.HandleFunc("/upload/", viewUpload)
	http.HandleFunc("/upload/up", upload)
	http.Handle("/upload/resources/", http.StripPrefix("/upload/", http.FileServer(http.FS(static))))

	println("ReceiveFile, Listening on port 10026")
	println("http://localhost:10026")
	http.ListenAndServe(":10026", nil)

}

func redirectToIndex(w http.ResponseWriter, req *http.Request) {

	checkIndexFile(req)
	http.Redirect(w, req, "/upload"+req.RequestURI, http.StatusTemporaryRedirect)
}
