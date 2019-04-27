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
		returnString += fmt.Sprintf("label: %s, probability: %.2f%%\n", l.Label, l.Probability*100)
	}

	_, err := w.Write([]byte(returnString))
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Println(err.Error())
		return
	}
}

func createMakeRequestHandler(computationPolicy middleware.ComputationPolicy) func(http.ResponseWriter, *http.Request) {
	client := middleware.MakePolicyAwareClient(computationPolicy)

	return func(w http.ResponseWriter, r *http.Request) {
		// Get local/remote preference from request
		requestParams := r.URL.Query()
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

		var httpRequest *http.Request
		if DOCKER {
			httpRequest, _ = http.NewRequest("PUT", "http://server:3000/", bytes.NewReader(body))
		} else {
			httpRequest, _ = http.NewRequest("PUT", "http://127.0.0.1:3000/", bytes.NewReader(body))

		}

		policy := middleware.RequestPolicy{
			RequesterID:                 "client1",
			PreferredProcessingLocation: preferredLocation,
			HasAllRequiredData:          true,
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

		// Check that computation level is correct
		if pamResp.ComputationLevel != middleware.CanCompute {
			http.Error(w, "Server could not perform object classification", 500)
			return
		}

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

func createUpdatePolicyHandler(computationPolicy *middleware.DynamicComputationPolicy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestParams := r.URL.Query()
		action := requestParams.Get("action")
		switch action {
		case "activate":
			err := computationPolicy.Activate("/", middleware.CanCompute)
			if err != nil {
				log.Printf(err.Error())
				http.Error(w, err.Error(), 500)
				return
			}
		case "deactivate":
			err := computationPolicy.Deactivate("/", middleware.CanCompute)
			if err != nil {
				log.Printf(err.Error())
				http.Error(w, err.Error(), 500)
				return
			}
		default:
			log.Println("No action was specified")
			_, err := w.Write([]byte("No action was specified"))
			if err != nil {
				log.Printf(err.Error())
				http.Error(w, err.Error(), 500)
				return
			}
		}

		_, err := w.Write([]byte("OK"))
		if err != nil {
			log.Printf(err.Error())
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

	computationPolicy := middleware.NewDynamicComputationPolicy()
	computationPolicy.Register("/", middleware.CanCompute, http.HandlerFunc(imageProcessingHandler))

	// Listen on 4000 for request to start example
	http.Handle("/request", http.HandlerFunc(createMakeRequestHandler(computationPolicy)))

	// Listen on 4000 for request to edit the computation policy
	http.Handle("/update-policy", http.HandlerFunc(createUpdatePolicyHandler(computationPolicy)))

	log.Println("Listening...")
	err := http.ListenAndServe(":4000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
