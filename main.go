package main

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"path/filepath"
	"url-shortener/middleware"

	"github.com/catinello/base62"
)

type PageData struct{
    Name string
    ShortenedUrl string
    OriginalUrl string
    Error string
}

var currentSuffix string
var generatedCode int = 100
var originalUrl string

var tmpl = template.Must(template.New("").ParseGlob("./templates/*")) 

func main() {
    router := http.NewServeMux()

    fileServer := http.FileServer(http.Dir("images"))
    router.Handle("GET /images/{file...}", fileServer)
   
    router.HandleFunc("GET /favicon.ico", faviconHandler)

    router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request){
        tmpl.ExecuteTemplate(w, "index.html", PageData{
            Name: "Url Shortener",
        })
    })

    router.HandleFunc("POST /shorten", shortenHandler)

    router.HandleFunc("GET /{suffix}", func(w http.ResponseWriter, r *http.Request) {
        _, err := url.ParseRequestURI(originalUrl)

        if err != nil {
            w.WriteHeader(http.StatusNotFound)
            tmpl.ExecuteTemplate(w, "error.html", PageData{
                Error: "404 Not Found",
            })
            return
        }
        http.Redirect(w, r, originalUrl, http.StatusFound)
    })

    srv := http.Server{
        Addr: ":9090",
        Handler: middleware.Logging(router),
    }

    fmt.Println("Starting website")
    err := srv.ListenAndServe()
    if err != nil && !errors.Is(err, http.ErrServerClosed) {
        fmt.Println("Error occured: ", err)
    }
}


func shortenHandler(w http.ResponseWriter, r *http.Request) {
    originalUrl = r.FormValue("url")
    
    currentSuffix = generateUrlSuffix()

    shortenedUrl := fmt.Sprintf("%s/%s",r.Host, currentSuffix)

    data := PageData{
        ShortenedUrl: shortenedUrl,
        OriginalUrl: originalUrl,

    }
    tmpl.ExecuteTemplate(w, "shorten.html", data)

}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "image/x-icon")
    http.ServeFile(w, r, filepath.Join("images", "favicon.ico"))
}

func generateUrlSuffix() string {
    encodedSuffix := base62.Encode(generatedCode)
    generatedCode += 10
    return encodedSuffix 
}
