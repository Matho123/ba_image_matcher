package image_service

import (
	"database/sql"
	"log"
)

const maxChunkSize = 50

func applyDatabaseOperation(applyFunction func(databaseConnection *sql.DB)) error {
	databaseConnection, err := openDatabaseConnection()
	if err != nil {
		return err
	}
	defer databaseConnection.Close()

	applyFunction(databaseConnection)

	return nil
}

func applyChunkedFeatureBasedRetrievalOperation(
	applyFunction func(databaseImage FeatureImageEntity), descriptor string,
) error {
	err := applyDatabaseOperation(func(databaseConnection *sql.DB) {
		offset := 0
		for {
			databaseImageChunk, err := retrieveFeatureImageChunk(
				databaseConnection,
				descriptor,
				offset,
				maxChunkSize+1,
			)
			if err != nil {
				log.Println("Error while retrieving chunk from database images: ", err)
			}

			for _, databaseImage := range (*databaseImageChunk)[0 : len(*databaseImageChunk)-1] {
				applyFunction(databaseImage)
			}

			if len(*databaseImageChunk) < maxChunkSize+1 {
				break
			}
			offset += maxChunkSize

			databaseImageChunk = nil
		}
	})
	return err
}

func applyChunkedPHashRetrievalOperation(applyFunction func(databaseImage PHashImageEntity)) error {
	err := applyDatabaseOperation(func(databaseConnection *sql.DB) {
		offset := 0
		for {
			databaseImageChunk, err := retrievePHashImageChunk(databaseConnection, offset, maxChunkSize+1)
			if err != nil {
				log.Println("Error while retrieving chunk from database images: ", err)
			}

			for _, databaseImage := range (*databaseImageChunk)[0 : len(*databaseImageChunk)-1] {
				applyFunction(databaseImage)
			}

			if len(*databaseImageChunk) < maxChunkSize+1 {
				break
			}
			offset += maxChunkSize
		}
	})
	return err
}
