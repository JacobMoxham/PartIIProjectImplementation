package server

import (
	"fmt"
	. "github.com/JacobMoxham/PartIIProjectImplementation/image_recognition"
	"github.com/JacobMoxham/PartIIProjectImplementation/middleware"
	"github.com/tensorflow/tensorflow/tensorflow/go"
	"log"
	"net/http"
)

func imageProcessingHandler(w http.ResponseWriter, r *http.Request) {
	modelGraph, labels, err := LoadModel()
	if err != nil {
		log.Fatalf("unable to load model: %v", err)
	}

	// Get normalized tensor
	tensor, err := NormalizeImage(r.Body)
	if err != nil {
		log.Fatalf("unable to make a tensor from image: %v", err)
	}

	// Create a session for inference over modelGraph
	session, err := tensorflow.NewSession(modelGraph, nil)
	if err != nil {
		log.Fatalf("could not init session: %v", err)
	}

	output, err := session.Run(
		map[tensorflow.Output]*tensorflow.Tensor{
			modelGraph.Operation("input").Output(0): tensor,
		},
		[]tensorflow.Output{
			modelGraph.Operation("output").Output(0),
		},
		nil)
	if err != nil {
		log.Fatalf("could not run inference: %v", err)
	}

	res := GetTopFiveLabels(labels, output[0].Value().([][]float32)[0])
	returnString := ""
	for _, l := range res {
		fmt.Printf("label: %s, probability: %.2f%%\n", l.Label, l.Probability*100)
		returnString += fmt.Sprintf("label: %s, probability: %.2f%%\n", l.Label, l.Probability*100)
	}

	_, err = w.Write([]byte(returnString))
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Println(err.Error())
		return
	}
}

func main() {
	// Create actual function to run
	finalHandler := http.HandlerFunc(imageProcessingHandler)

	// Define computation policy
	computationPolicy := middleware.NewStaticComputationPolicy()
	computationPolicy.Register("/", middleware.CanCompute, finalHandler)

	// Chain together "other" middlewares
	handler := middleware.PrivacyAwareHandler(computationPolicy)

	// Register the composite handler at '/' on port 3000
	http.Handle("/", handler)
	log.Println("Listening...")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
