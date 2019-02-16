package main

import (
	"bytes"
	"fmt"
	imageProcessing "github.com/JacobMoxham/PartIIProjectImplementation/image_recognition"
	"github.com/JacobMoxham/PartIIProjectImplementation/middleware"
	image2 "image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const DOCKER = true

func imageProcessingHandler(w http.ResponseWriter, r *http.Request) {
	top5Labels := imageProcessing.GetTop5LabelsFromImageReader(r.Body)
	returnString := ""
	for _, l := range top5Labels {
		fmt.Printf("label: %s, probability: %.2f%%\n", l.Label, l.Probability*100)
		returnString += fmt.Sprintf("label: %s, probability: %.2f%%\n", l.Label, l.Probability*100)
	}

	_, err := w.Write([]byte(returnString))
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Println(err.Error())
		return
	}
}

func createMakeRequestHandler() func(http.ResponseWriter, *http.Request) {
	computationPolicy := middleware.NewStaticComputationPolicy()
	computationPolicy.Register("/", middleware.CanCompute, http.HandlerFunc(imageProcessingHandler))

	client := middleware.MakePrivacyAwareClient(computationPolicy)

	return func(w http.ResponseWriter, r *http.Request) {
		policy := middleware.RequestPolicy{
			RequesterID:                 "client1",
			PreferredProcessingLocation: middleware.Remote,
			HasAllRequiredData:          true,
		}

		pwd, _ := os.Getwd()
		filePath := "images/image1.jpg"
		if !DOCKER {
			filePath = "usecases/image_processing/client/images/image1.jpg"
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

		// TODO: get image from "phone"
		buf := new(bytes.Buffer)
		err = jpeg.Encode(buf, image, nil)
		body := buf.Bytes()

		var httpRequest *http.Request
		if DOCKER {
			httpRequest, _ = http.NewRequest("PUT", "http://server:3000/", bytes.NewReader(body))
		} else {
			httpRequest, _ = http.NewRequest("PUT", "http://127.0.0.1:3000/", bytes.NewReader(body))

		}

		req := middleware.PamRequest{
			Policy:      &policy,
			HttpRequest: httpRequest,
		}

		pamResp, err := client.Send(req)
		if err != nil {
			log.Println("Error:", err)
			return
		}
		if pamResp.HttpResponse.StatusCode < 200 || pamResp.HttpResponse.StatusCode >= 300 {
			log.Printf("Request produced a none 2xx status code: %s", pamResp.HttpResponse.Status)
			return
		}

		// TODO: check that computation level is correct

		resp := pamResp.HttpResponse

		// Read response
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal("Error reading body of response.", err)
		}
		resp.Body.Close()

		log.Println(fmt.Sprintf("Code: %d Result: %s", resp.StatusCode, body))
		_, err = w.Write([]byte(body))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
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
