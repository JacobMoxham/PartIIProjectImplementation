package main

import (
	"fmt"
	"github.com/JacobMoxham/PartIIProjectImplementation/middleware"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	policy := middleware.RequestPolicy{
		RequesterID:                 "server",
		PreferredProcessingLocation: middleware.Remote,
		HasAllRequiredData:          false,
	}
	httpRequest, _ := http.NewRequest("GET", "http://127.0.0.1:3001/", nil)
	req := middleware.PamRequest{
		Policy:      &policy,
		HttpRequest: httpRequest,
	}
	resp, err := req.Send()
	if err != nil {
		log.Println("Error:", err)
		return
	}

	// Read response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading body of response.", err)
	}
	resp.Body.Close()

	log.Println(fmt.Sprintf("%s", body))
}