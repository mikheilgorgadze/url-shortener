package main

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"time"
	"url-shortener/database"
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
var generatedCode int = 100000
const INCREMENT = 50000
//var originalUrl string

var tmpl = template.Must(template.New("").ParseGlob("./templates/*")) 

func main() {
    var err error
    router := http.NewServeMux()
    err = database.InitDB()
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    
    defer database.CloseDB()

    fileServer := http.FileServer(http.Dir("images"))
    router.Handle("GET /images/{file...}", fileServer)
   
    router.HandleFunc("GET /favicon.ico", faviconHandler)

    router.HandleFunc("GET /", indexPageHandler)

    router.HandleFunc("POST /shorten", shortenHandler)

    router.HandleFunc("GET /{suffix}", urlRedirectHandler)

    srv := http.Server{
        Addr: ":9090",
        Handler: middleware.Logging(router),
    }

    fmt.Println("Starting website")
    err = srv.ListenAndServe()
    if err != nil && !errors.Is(err, http.ErrServerClosed) {
        fmt.Println("Error occured: ", err)
    }

    
}


func shortenHandler(w http.ResponseWriter, r *http.Request) {
    originalUrl := r.FormValue("url")

    // validate url
    if _, err := url.ParseRequestURI(originalUrl); err != nil {
        data := PageData {
            Error: "Invalid URL provided",
        }
        tmpl.ExecuteTemplate(w, "error.html", data)
        return
    }

    currentSuffix, err := generateUniqueShortenCode(); 
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        tmpl.ExecuteTemplate(w, "error.html", PageData{
            Error: "Something unexpected happened",
        })
        return
    }
    shortenedUrl := fmt.Sprintf("%s/%s",r.Host, currentSuffix)

    generatedUrl := database.GeneratedUrl{
        ShortCode: currentSuffix,
        LongUrl: originalUrl,
        AddedTime: time.Now(),
    }
    
    // save url
    err = database.InsertURL(generatedUrl)
    if err != nil {
        log.Printf("Error storing URL: %v", err)
        data := PageData {
            Error: "Failed to create shortened URL",
        }
        tmpl.ExecuteTemplate(w, "error.html", data)
        return
    }

    data := PageData{
        ShortenedUrl: shortenedUrl,
        OriginalUrl: originalUrl,
    }

    tmpl.ExecuteTemplate(w, "shorten.html", data)

}

func indexPageHandler(w http.ResponseWriter, r *http.Request){
        tmpl.ExecuteTemplate(w, "index.html", PageData{
            Name: "Url Shortener",
        })
}

func urlRedirectHandler(w http.ResponseWriter, r *http.Request) {
    shortCode := r.PathValue("suffix")
    originalUrl, err := database.GetURLByShortCode(shortCode)

    if err != nil {
        w.WriteHeader(http.StatusNotFound)
        tmpl.ExecuteTemplate(w, "error.html", PageData{
            Error: "404 Not Found",
        })
        return
    }
   // _, err := url.ParseRequestURI(originalUrl)

    http.Redirect(w, r, originalUrl, http.StatusFound)
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "image/x-icon")
    http.ServeFile(w, r, filepath.Join("images", "favicon.ico"))
}

func generateUniqueShortenCode() (string, error) {
    maxRetry := 10
    for retry := 0; retry < maxRetry; retry ++{
        encodedSuffix := base62.Encode(generatedCode)
        exists, err := database.ShortCodeExists(encodedSuffix)
        if err != nil {
            log.Printf("Error checking shortCode: %v", err)
            generatedCode += INCREMENT 
            continue
        }

        if !exists {
            generatedCode += INCREMENT 
            return encodedSuffix, nil
        }
        generatedCode += INCREMENT 
    }
    return "", fmt.Errorf("failed to generate unique code after %d attempts", maxRetry)
}
