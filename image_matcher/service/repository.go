package service

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

type ImageEntity struct {
	id                int
	externalReference string
	descriptorData    []byte
}

func openDatabaseConnection() (*sql.DB, error) {
	databaseConnection, err := sql.Open("mysql", "root:root@/images")

	if err != nil {
		return nil, errors.New(fmt.Sprintf("couldn't connect to database %s", err.Error()))
	}
	databaseConnection.SetConnMaxLifetime(-1)
	databaseConnection.SetMaxOpenConns(10)
	databaseConnection.SetMaxIdleConns(10)

	return databaseConnection, nil
}

func insertImageIntoDatabase(databaseConnection *sql.DB, processedImageDTO ProcessedImage) error {
	externalReference := processedImageDTO.externalReference
	descriptorData := processedImageDTO.descriptorData

	_, err := databaseConnection.Exec(
		"INSERT INTO image (external_reference, descriptor_data) VALUES (?, ?)",
		externalReference,
		descriptorData,
	)

	if err != nil {
		return errors.New(fmt.Sprintf("couldn't insert %s into database %s", externalReference, err.Error()))
	}
	log.Println(fmt.Sprintf("Inserted %s into Database", externalReference))
	return nil
}

func retrieveImageChunkFromDatabase(databaseConnection *sql.DB, offset int, limit int) ([]ImageEntity, error) {
	imageRows, err := databaseConnection.Query(
		"SELECT id, external_reference, descriptor_data"+
			" FROM image"+
			" LIMIT ? OFFSET ?",
		limit,
		offset,
	)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("couldn't retreive images from database: %s", err.Error()))
	}
	defer imageRows.Close()

	var imageEntityChunk []ImageEntity

	for imageRows.Next() {
		var id int
		var externalReference string
		var descriptorData []byte

		err := imageRows.Scan(&id, &externalReference, &descriptorData)

		if err != nil {
			return nil, err
		}

		imageEntityChunk = append(
			imageEntityChunk,
			ImageEntity{
				id:                id,
				externalReference: externalReference,
				descriptorData:    descriptorData,
			},
		)

		log.Println(fmt.Sprintf("Retrieved %s from Database", externalReference))
	}
	return imageEntityChunk, nil
}
