package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func (s *Server) handleGetCardImage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//get entity id from url
		entityID := r.URL.Query().Get("entityid")

		//check if a entityid has been provided
		if entityID == "" {
			//Not set, send 400 bad request
			w.WriteHeader(400)
			fmt.Fprint(w, "No entity ID provided.")
			fmt.Println("No entity ID has been provided.")
			return
		}

		//print to the console that a file was requested
		fmt.Println("Card Image has been requested for entity --> " + entityID)

		//Get file names from CRUD
		req, respErr := http.Get("http://" + conf.CRUDHost + ":" + conf.CRUDPort + "/cardimage?entityid=" + entityID)

		//check for response error of 500
		if respErr != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, respErr.Error())
			fmt.Println("Error in communication with CRUD service endpoint for request to get file details")
			return
		}
		if req.StatusCode != 200 {
			fmt.Println("Request to DB can't be completed to get file details")
		}
		if req.StatusCode == 500 {
			w.WriteHeader(500)
			bodyBytes, err := ioutil.ReadAll(req.Body)
			if err != nil {
				log.Fatal(err)
			}
			bodyString := string(bodyBytes)
			fmt.Fprintf(w, "Database error occured upon retrieval"+bodyString)
			fmt.Println("Database error occured upon retrieval" + bodyString)
			return
		}

		//close the request
		defer req.Body.Close()

		//create new response struct
		var cardImage CardImage
		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(&cardImage)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err.Error())
			fmt.Println("Unable to decode cardImage response")
			return
		}

		//check if the file exists in the file system
		fileName := conf.ResourcesPath + entityID + "/" + cardImage.FilePath

		if cardImage.FilePath == "" {
			w.WriteHeader(500)
			fmt.Fprint(w, "Image not found -->"+fileName)
			fmt.Println("No image found")
			return
		}

		Openfile, err := os.Open(fileName)
		defer Openfile.Close()

		if err != nil {
			//File not found
			w.WriteHeader(400)
			fmt.Fprint(w, "File was not found in file system")
			fmt.Println("File not found: " + fileName)
			return
		}

		FileHeader := make([]byte, 512)

		Openfile.Read(FileHeader)

		//We already read 512 bytes, so we reset the offset back to 0
		Openfile.Seek(0, 0)
		/*
			THE BELOW WAS FOR WHEN WE WERE SENDING THE FILE BACK AS A FILE
			FileContentType := http.DetectContentType(FileHeader)
			//Get info from file
			FileStat, _ := Openfile.Stat()
			//get size of the file as a string value
			FileSize := strconv.FormatInt(FileStat.Size(), 10)
			//Send the headers
			w.Header().Set("Content-Disposition", "attachment; filename="+cardImage.FileName)
			w.Header().Set("Content-Type", FileContentType)
			w.Header().Set("Content-Length", FileSize)
			//Send the file
			//io.Copy(w, Openfile)
		*/

		bytes, err := ioutil.ReadAll(Openfile)
		if err != nil {
			log.Fatal(err)
		}

		cardimagebytes := CardImageBytes{}
		cardimagebytes.EntityID = entityID
		cardimagebytes.ImageBytes = bytes

		fmt.Println("Image converted to []byte and sent to caller for entityid --> " + entityID)
		// converting response struct to JSON payload to send to service that called this function.
		js, jserr := json.Marshal(cardimagebytes)

		// check to see if any errors occured with coverting to JSON.
		if jserr != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "Unable to create JSON object from DB result to fetch Card Image")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(js)
	}
}

func (s *Server) handlePostCardImageBatch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Handle post Card Image Path Batch has been called...")

		//get JSON payload
		imagerequest := CardImageBatchRequest{}
		err := json.NewDecoder(r.Body).Decode(&imagerequest)

		//handle for bad JSON provided
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "Bad JSON provided to get batch image paths from CRUD")
			return
		}

		requestByte, _ := json.Marshal(imagerequest)
		req, respErr := http.Post("http://"+conf.CRUDHost+":"+conf.CRUDPort+"/cardimagebatch", "application/json", bytes.NewBuffer(requestByte))

		//check for response error of 500
		if respErr != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, respErr.Error())
			fmt.Println("Error in communication with CRUD service endpoint for request to get file details")
			return
		}
		if req.StatusCode != 200 {
			fmt.Println("Request to DB can't be completed to get file details")
		}
		if req.StatusCode == 500 {
			w.WriteHeader(500)
			bodyBytes, err := ioutil.ReadAll(req.Body)
			if err != nil {
				log.Fatal(err)
			}
			bodyString := string(bodyBytes)
			fmt.Fprintf(w, "Database error occured upon retrieval"+bodyString)
			fmt.Println("Database error occured upon retrieval" + bodyString)
			return
		}

		//close the request
		defer req.Body.Close()

		//create new response struct
		cardImageBatch := CardImageBatch{}
		decoder := json.NewDecoder(req.Body)
		err = decoder.Decode(&cardImageBatch)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err.Error())
			fmt.Println("Unable to decode cardImage response")
			return
		}

		cardimages := CardBytesBatch{}
		cardimages.Images = []CardImageBytes{}

		//get bytes for all images returned by the crud
		for _, image := range cardImageBatch.Images {
			//check if the file exists in the file system
			fileName := conf.ResourcesPath + image.EntityID + "/" + image.FilePath
			if image.FilePath == "" {
				w.WriteHeader(500)
				fmt.Fprint(w, "Image not found -->"+fileName)
				fmt.Println("No image found")
				return
			}

			Openfile, err := os.Open(fileName)
			defer Openfile.Close()

			if err != nil {
				//File not found
				w.WriteHeader(400)
				fmt.Fprint(w, "File was not found in file system")
				fmt.Println("File not found: " + fileName)
				return
			}

			FileHeader := make([]byte, 512)
			Openfile.Read(FileHeader)
			//We already read 512 bytes, so we reset the offset back to 0
			Openfile.Seek(0, 0)
			bytes, err := ioutil.ReadAll(Openfile)
			if err != nil {
				log.Fatal(err)
			}
			cardimagebytes := CardImageBytes{}
			cardimagebytes.EntityID = image.EntityID
			cardimagebytes.ImageBytes = bytes

			fmt.Println("Image converted to []byte and sent to caller for entityid --> " + image.EntityID)
			cardimages.Images = append(cardimages.Images, cardimagebytes)
		}

		// converting response struct to JSON payload to send to service that called this function.
		js, jserr := json.Marshal(cardimages)

		// check to see if any errors occured with coverting to JSON.
		if jserr != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "Unable to create JSON object from DB result to fetch Card Images")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(js)
	}
}
