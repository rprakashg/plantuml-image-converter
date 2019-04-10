package main

import (
	"context"
	"flag"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"go.uber.org/zap"
)

const (
	//ContentTypePNG represents png content type
	ContentTypePNG = "image/png"
	TempDir        = "/var/tmp"
)

var (
	addr        = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
	wait        time.Duration
	srv         *http.Server
	plantUMLJar string
)

func init() {
	plantUMLJar = os.Getenv("PLANTUML_JAR")
}

func main() {
	r := mux.NewRouter()
	r.Methods("POST").
		Path("/").
		Name("Home").
		HandlerFunc(homeHandler)

	srv = &http.Server{
		Addr:         *addr,
		WriteTimeout: time.Second * 10,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}

	go func() {
		log.Debugf("Server listening on port %s \n", *addr)
		if err := srv.ListenAndServe(); err != nil {
			log.Errorf("Error starting the server %v", zap.Any("exception", err))
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()

	srv.Shutdown(ctx)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	var fileName string

	//read the payload from request
	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("uploadFile")
	if err != nil {
		log.Errorf("Error getting Multipart form data")
		http.Error(w, "No data posted", http.StatusBadRequest)
	}
	defer file.Close()
	f, err := os.OpenFile(TempDir+"/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Errorf("Error saving file to temp")
		http.Error(w, "Internal server error, please try again....", http.StatusInternalServerError)
	}
	defer f.Close()
	io.Copy(f, file)
	//read headers
	acceptHeader := r.Header.Get("Accept")
	if len(acceptHeader) == 0 {
		log.Error("Accept header is required")
		http.Error(w, "Accept header is required", http.StatusBadRequest)
	}

	switch acceptHeader {
	case ContentTypePNG:
		log.Infof("PlantUML Jar: %s", plantUMLJar)
		fileName = TempDir + "/" + handler.Filename

		//convert
		cmd := exec.Command("java", "-jar", plantUMLJar, fileName)
		err := cmd.Run()
		if err != nil {
			log.Errorf("Error generating PNG %v", zap.Any("error", err))
			http.Error(w, "Error generating PNG", http.StatusInternalServerError)
		}
		i := strings.Index(fileName, ".")
		pngFile := fileName[0:i]
		pngFile = pngFile + ".png"
		log.Infof("PNG file name: %s", pngFile)
		//read file and write to output
		png, err := ioutil.ReadFile(pngFile)
		if err != nil {
			log.Error("Error reading png file from temp")
			http.Error(w, "Internal server error, please try again...", http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", ContentTypePNG)
		w.WriteHeader(http.StatusOK)
		w.Write(png)

		break
	default:
		http.Error(w, "Invalid accept header", http.StatusBadRequest)
	}
}
