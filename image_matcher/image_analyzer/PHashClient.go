package image_analyzer

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

type PHashDTO struct {
	Hash string `json:"hash"`
	//runtime in seconds
	Runtime float64 `json:"runtime"`
}

const phashCalculatorUrl = "http://localhost:8000"

func GetPHashValue(image *image.Image) (uint64, time.Duration) {
	if !isURLUp(phashCalculatorUrl + "/status") {
		log.Println("using local phash implementation!")
		return CalculateHash(image)
	}

	imageByteBuffer := new(bytes.Buffer)
	err := png.Encode(imageByteBuffer, *image)
	if err != nil {
		log.Fatal("couldn't create bytebuffer from image!")
	}

	response, err := http.Post(phashCalculatorUrl+"/calculateHash", "application/json", imageByteBuffer)
	if err != nil {
		log.Fatal("failed to do request: ", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Fatalf("HTTP request failed with status code %d", response.StatusCode)
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal("Error while sending request: ", err)
	}

	var hashDTO PHashDTO
	err = json.Unmarshal(responseBody, &hashDTO)
	if err != nil {
		log.Fatal("Error while trying to unmarshal responsebody: ", err)
	}

	uIntHash, err := strconv.ParseUint(hashDTO.Hash, 16, 64)
	if err != nil {
		log.Fatal("Error while converting hash to uint: ", err)
	}
	return uIntHash, time.Duration(hashDTO.Runtime * float64(time.Second))
}

func isURLUp(url string) bool {
	response, err := http.Get(url)
	if err != nil {
		return false
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusOK {
		return true
	}

	return false
}
