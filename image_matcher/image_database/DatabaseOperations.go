package image_database

import (
	"database/sql"
	"log"
)

const MaxChunkSize = 50

func ApplyDatabaseOperation(applyFunction func(databaseConnection *sql.DB)) error {
	databaseConnection, err := openDatabaseConnection()
	if err != nil {
		return err
	}
	defer databaseConnection.Close()

	applyFunction(databaseConnection)

	return nil
}

func ApplyChunkedFeatureBasedRetrievalOperation(
	applyFunction func(databaseImage FeatureImageEntity), descriptor string,
) error {
	err := ApplyDatabaseOperation(func(databaseConnection *sql.DB) {
		offset := 0
		for {
			databaseImageChunk, err := retrieveFeatureImageChunk(
				databaseConnection,
				descriptor,
				offset,
				MaxChunkSize+1,
			)
			if err != nil {
				log.Println("Error while retrieving chunk from database images: ", err)
			}

			for _, databaseImage := range (*databaseImageChunk)[0 : len(*databaseImageChunk)-1] {
				applyFunction(databaseImage)
			}

			if len(*databaseImageChunk) < MaxChunkSize+1 {
				break
			}
			offset += MaxChunkSize

			databaseImageChunk = nil
		}
	})
	return err
}

func ApplyChunkedPHashRetrievalOperation(applyFunction func(databaseImage PHashImageEntity)) error {
	err := ApplyDatabaseOperation(func(databaseConnection *sql.DB) {
		offset := 0
		for {
			databaseImageChunk, err := retrievePHashImageChunk(databaseConnection, offset, MaxChunkSize+1)
			if err != nil {
				log.Println("Error while retrieving chunk from database images: ", err)
			}

			for _, databaseImage := range (*databaseImageChunk)[0 : len(*databaseImageChunk)-1] {
				applyFunction(databaseImage)
			}

			if len(*databaseImageChunk) < MaxChunkSize+1 {
				break
			}
			offset += MaxChunkSize
		}
	})
	return err
}

func ApplyChunkedHybridRetrievalOperation(applyFunction func(databaseImage HybridEntity)) error {
	err := ApplyDatabaseOperation(func(databaseConnection *sql.DB) {
		offset := 0
		for {
			databaseImageChunk, err := retrieveHybridChunk(databaseConnection, offset, MaxChunkSize+1)
			if err != nil {
				log.Println("Error while retrieving chunk from database images: ", err)
			}

			for _, databaseImage := range (*databaseImageChunk)[0 : len(*databaseImageChunk)-1] {
				applyFunction(databaseImage)
			}

			if len(*databaseImageChunk) < MaxChunkSize+1 {
				break
			}
			offset += MaxChunkSize
		}
	})
	return err
}
