package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
)

//This function returns one image, for the specific advertisement requested, only returns the main image
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

		cardimagebytes := CardImageBytes{}
		cardimagebytes.EntityID = entityID
		if cardImage.FileName != "" {
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

			cardimagebytes.ImageBytes = bytes

			fmt.Println("Image converted to []byte and sent to caller for entityid --> " + entityID)
		} else {
			//Get bytes for default image
			defaultFileName := conf.ResourcesPath + "default/default.png"
			Openfile, err := os.Open(defaultFileName)
			defer Openfile.Close()

			if err != nil {
				//File not found
				w.WriteHeader(400)
				fmt.Fprint(w, "Default file was not found in file system")
				fmt.Println("Default file not found: " + defaultFileName)
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
			cardimagebytes.ImageBytes = bytes
		}

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

//This function returns multiple images, based on all the advertisements requested... One image per advertisement
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

		//Get bytes for default image
		defaultFileName := conf.ResourcesPath + "default/default.png"
		Openfile, err := os.Open(defaultFileName)
		defer Openfile.Close()

		if err != nil {
			//File not found
			w.WriteHeader(400)
			fmt.Fprint(w, "Default file was not found in file system")
			fmt.Println("Default file not found: " + defaultFileName)
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
		defaultImageBytes := bytes
		//Check if all entities supplied by caller have an image attached
		imageExists := false
		for _, entity := range imagerequest.Cards {
			for _, result := range cardimages.Images {
				if entity.EntityID == result.EntityID {
					imageExists = true
				}
			}
			//If no result was returned from the db, provide default image as result
			if !imageExists {
				cardimagebytes := CardImageBytes{}
				cardimagebytes.EntityID = entity.EntityID
				cardimagebytes.ImageBytes = defaultImageBytes
				cardimages.Images = append(cardimages.Images, cardimagebytes)
				fmt.Println("Default image converted to []byte and sent to caller for entityid --> " + entity.EntityID)

			}
			imageExists = false
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

//This function returns all images for a specific advertisement
func (s *Server) handleGetAdvertisementImages() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Handle post Card Image Path Batch has been called...")

		//get advert id from url
		advertisementid := r.URL.Query().Get("advertisementid")

		//check if a entityid has been provided
		if advertisementid == "" {
			//Not set, send 400 bad request
			w.WriteHeader(400)
			fmt.Fprint(w, "No advertisement ID provided.")
			fmt.Println("No advertisement ID has been provided.")
			return
		}

		//print to the console that a file was requested
		fmt.Println("All images have been requested for advertisement --> " + advertisementid)

		//Get file names from CRUD
		req, respErr := http.Get("http://" + conf.CRUDHost + ":" + conf.CRUDPort + "/advertisementimages?advertisementid=" + advertisementid)

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
		err := decoder.Decode(&cardImageBatch)
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
			//If no image exists, get the default image

			if image.FilePath == "default" {
				fileName = conf.ResourcesPath + "default/default.png"
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

func (s *Server) handleUploadImage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Handle Upload Image has been called...")

		//get JSON payload
		image := UploadImage{}
		err := json.NewDecoder(r.Body).Decode(&image)

		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err.Error())
			fmt.Println("Unable to decode JSON...")
			return
		}

		//generates a new uuid
		newid := uuid.New()
		//generates the new path for the image
		newpath := conf.ResourcesPath + image.EntityID + "/" + newid.String()

		//Make directory
		if _, err := os.Stat(conf.ResourcesPath + image.EntityID); os.IsNotExist(err) {
			os.Mkdir(conf.ResourcesPath+image.EntityID, os.ModeDir)
		}
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err.Error())
			fmt.Println("Unable to create new directory in file system...")
			return
		}
		//0644 = permissions required
		err = ioutil.WriteFile(newpath, image.ImageBytes, 0644)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err.Error())
			fmt.Println("Unable to write file to file system...")
			return
		}

		imagePost := UploadImageInformation{}
		imagePost.EntityID = image.EntityID
		imagePost.FileName = "UploadedImg"
		imagePost.FilePath = newid.String()
		imagePost.IsMainImage = image.IsMainImage

		//write file data to db
		requestByte, _ := json.Marshal(imagePost)
		req, respErr := http.Post("http://"+conf.CRUDHost+":"+conf.CRUDPort+"/uploadimage", "application/json", bytes.NewBuffer(requestByte))

		if respErr != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, respErr.Error())
			fmt.Println("Error in communication with CRUD service endpoint for request insert image details")
			return
		}
		if req.StatusCode != 200 {
			fmt.Println("Request to DB can't be completed to insert image details")
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

		w.WriteHeader(200)
		fmt.Fprintf(w, "Image has been saved to the file system.")
		fmt.Println("Image written to file system for entity --> " + image.EntityID)
	}
}

func (s *Server) handleUploadImageBatch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Handle Upload Image has been called...")

		//get JSON payload
		images := UploadImageBatch{}
		err := json.NewDecoder(r.Body).Decode(&images)

		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err.Error())
			fmt.Println("Unable to decode JSON...")
			return
		}

		for _, image := range images.Images {
			//generates a new uuid
			newid := uuid.New()
			//generates the new path for the image
			newpath := conf.ResourcesPath + image.EntityID + "/" + newid.String()

			//Make directory
			if _, err := os.Stat(conf.ResourcesPath + image.EntityID); os.IsNotExist(err) {
				os.Mkdir(conf.ResourcesPath+image.EntityID, os.ModeDir)
			}
			if err != nil {
				w.WriteHeader(500)
				fmt.Fprint(w, err.Error())
				fmt.Println("Unable to create new directory in file system...")
				return
			}
			//0644 = permissions required
			err = ioutil.WriteFile(newpath, image.ImageBytes, 0644)
			if err != nil {
				w.WriteHeader(500)
				fmt.Fprint(w, err.Error())
				fmt.Println("Unable to write file to file system...")
				return
			}

			imagePost := UploadImageInformation{}
			imagePost.EntityID = image.EntityID
			imagePost.FileName = "UploadedImg"
			imagePost.FilePath = newid.String()
			imagePost.IsMainImage = image.IsMainImage

			//write file data to db
			requestByte, _ := json.Marshal(imagePost)
			req, respErr := http.Post("http://"+conf.CRUDHost+":"+conf.CRUDPort+"/uploadimage", "application/json", bytes.NewBuffer(requestByte))

			if respErr != nil {
				w.WriteHeader(500)
				fmt.Fprint(w, respErr.Error())
				fmt.Println("Error in communication with CRUD service endpoint for request insert image details")
				return
			}
			if req.StatusCode != 200 {
				fmt.Println("Request to DB can't be completed to insert image details")
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
		}

		w.WriteHeader(200)
		fmt.Fprintf(w, "Images have been saved to the file system.")
	}
}
