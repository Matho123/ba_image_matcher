package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type ImageList struct {
	Designs []Design `json:"designs"`
}

type Design struct {
	Id string `json:"id"`
}

func main() {
	targetDirectory := os.Args[1]
	if targetDirectory == "" {
		log.Fatal("no target directory selected")
	}

	directoryInfo, err := os.Stat(targetDirectory)
	if err != nil || !directoryInfo.IsDir() {
		log.Fatal("target directory is not a valid")
	}

	designs := GetDownloadableImageIds()

	for _, design := range designs {
		DownloadImageFromUrl(targetDirectory, design.Id)
		time.Sleep(3 * time.Second) //space out requests to not overload the server
	}
}

func GetDownloadableImageIds() []Design {
	endpoint := "https://ff.spod.com/fulfillment/public/api/designs?locale=en_US&platform=NA&query=&safeSearch=ALL&offset=0&limit=1000"
	response, err := http.Get(endpoint)
	if err != nil {
		fmt.Println("Error sending GET request:", err)
		return nil
	}
	defer response.Body.Close()

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response data:", err)
		return nil
	}

	var imageList ImageList
	err = json.Unmarshal(responseData, &imageList)
	if err != nil {
		fmt.Println("Error unmarshaling JSON: ", err)
		return nil
	}

	return imageList.Designs
}

func DownloadImageFromUrl(targetDirectory string, id string) {
	imageUrl := fmt.Sprintf("https://image.spreadshirtmedia.com/image-server/v1/designs/%s?width=1000", id)

	response, err := http.Get(imageUrl)
	if err != nil {
		fmt.Println("Error sending GET request:", err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		fmt.Println("HTTP GET request returned status:", response.Status)
		return
	}

	outputFile, err := os.Create(targetDirectory + "/" + id + ".png")
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	defer outputFile.Close()

	_, err = io.Copy(outputFile, response.Body)
	if err != nil {
		fmt.Println("Error copying image data:", err)
		return
	}

	fmt.Println("Image " + id + " downloaded successfully.")
}
