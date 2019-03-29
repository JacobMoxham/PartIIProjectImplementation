package main

import (
	"bytes"
	"fmt"
	imageProcessing "github.com/JacobMoxham/PartIIProjectImplementation/image_recognition"
	"github.com/JacobMoxham/PartIIProjectImplementation/middleware"
	"github.com/tcnksm/go-httpstat"
	image2 "image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const DOCKER = true

func processImage(imageBytes []byte) string {
	top5Labels := imageProcessing.GetTop5LabelsFromImageReader(ioutil.NopCloser(bytes.NewReader(imageBytes)))
	returnString := ""
	for _, l := range top5Labels {
		returnString += fmt.Sprintf("label: %s, probability: %.2f%%\n", l.Label, l.Probability*100)
	}

	return returnString
}

func createMakeRequestHandler() func(http.ResponseWriter, *http.Request) {
	client := http.Client{}

	return func(w http.ResponseWriter, r *http.Request) {
		// Get local/remote preference from request
		requestParams := r.URL.Query()

		// Get local/remote preference from request
		preferredLocationString := requestParams.Get("preferredLocation")
		preferredLocation, err := middleware.ProcessingLocationFromString(preferredLocationString)
		if err != nil {
			log.Println("Error:", err)
			return
		}

		imageFileName := requestParams.Get("imageFileName")
		if imageFileName == "" {
			imageFileName = "9kB.jpg"
			log.Println("Using 9kB.jpg as no imageFileName was specified")
		}

		pwd, _ := os.Getwd()
		filePath := "images/" + imageFileName
		if !DOCKER {
			filePath = "usecases/image_processing/client/images/" + imageFileName
		}

		imageFile, err := os.Open(filepath.Join(pwd, filePath))
		if err != nil {
			log.Println("Error:", err)
			return
		}
		defer imageFile.Close()

		image, _, err := image2.Decode(imageFile)
		if err != nil {
			log.Println("Error:", err)
			return
		}

		buf := new(bytes.Buffer)
		err = jpeg.Encode(buf, image, nil)
		body := buf.Bytes()

		if preferredLocation == middleware.Local {
			// Do the image processing locally
			_, err = w.Write([]byte(processImage(body)))
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			return
		}

		var httpRequest *http.Request
		if DOCKER {
			httpRequest, _ = http.NewRequest("PUT", "http://server:3000/", bytes.NewReader(body))
		} else {
			httpRequest, _ = http.NewRequest("PUT", "http://127.0.0.1:3000/", bytes.NewReader(body))

		}
		// Create a httpstat powered context
		var result httpstat.Result
		ctx := httpstat.WithHTTPStat(httpRequest.Context(), &result)
		httpRequest = httpRequest.WithContext(ctx)

		resp, err := client.Do(httpRequest)
		if err != nil {
			log.Println("Error:", err)
			return
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			log.Printf("Request produced a none 2xx status code: %s", resp.Status)
			return
		}

		// Read response
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal("Error reading body of response.", err)
		}
		resp.Body.Close()

		log.Println(fmt.Sprintf("Code: %d Result: %s", resp.StatusCode, body))
		_, err = w.Write(body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// Log the request latency from httpstat
		log.Printf("DNS lookup: %d ms", int(result.DNSLookup/time.Millisecond))
		log.Printf("TCP connection: %d ms", int(result.TCPConnection/time.Millisecond))
		log.Printf("TLS handshake: %d ms", int(result.TLSHandshake/time.Millisecond))
		log.Printf("Server processing: %d ms", int(result.ServerProcessing/time.Millisecond))
		log.Printf("Content transfer: %d ms", int(result.ContentTransfer(time.Now())/time.Millisecond))
	}
}

func main() {
	// Logging for performance analysis
	go func() {
		log.Println(http.ListenAndServe("0.0.0.0:6060", nil))
	}()

	// Listen on 4000 for request to start example
	http.Handle("/request", http.HandlerFunc(createMakeRequestHandler()))

	log.Println("Listening...")
	err := http.ListenAndServe(":4000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
