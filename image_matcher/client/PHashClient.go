package client

import (
	"bytes"
	"encoding/json"
	"image"
	"image/jpeg"
	"io"
	"log"
	"net/http"
	"strconv"
)

func GetPHashValue(image image.Image) uint64 {
	imageByteBuffer := new(bytes.Buffer)
	err := jpeg.Encode(imageByteBuffer, image, nil)
	if err != nil {
		log.Fatal("couldn't create bytebuffer from image!")
	}

	response, err := http.Post("http://localhost:8000/calculateHash", "application/json", imageByteBuffer)
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

	var hashDTO PHash
	err = json.Unmarshal(responseBody, &hashDTO)
	if err != nil {
		log.Fatal("Error while trying to unmarshal responsebody: ", err)
	}

	uIntHash, err := strconv.ParseUint(hashDTO.Hash, 16, 64)
	if err != nil {
		log.Fatal("Error while converting hash to uint: ", err)
	}
	//println(hashDTO.Hash)
	//println(strconv.FormatUint(uIntHash, 2))

	return uIntHash
}

type PHash struct {
	Hash string `json:"hash"`
}
