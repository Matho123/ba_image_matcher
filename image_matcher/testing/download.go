package testing

import (
	"encoding/json"
	"fmt"
	"io"
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

func downloadOriginalImages([]string) {
	designs := getDownloadableImageIds()

	for _, design := range designs {
		downloadImageFromUrl(design.Id)
		time.Sleep(5 * time.Second)
	}
}

func getDownloadableImageIds() []Design {
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

func downloadImageFromUrl(id string) {
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

	outputFile, err := os.Create("images/originals/" + id + ".png")
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
