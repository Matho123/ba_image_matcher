package service

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

type ForbiddenImageCreation struct {
	externalReference string
	siftDescriptor    []byte
	orbDescriptor     []byte
	briskDescriptor   []byte
	pHash             uint64
}

type SearchImageCreation struct {
	externalReference string
	originalReference string
	scenario          string
	modificationInfo  string
}

type FeatureImageEntity struct {
	externalReference string
	descriptors       []byte
}

type PHashImageEntity struct {
	externalReference string
	hash              uint64
}

type SearchImageEntity struct {
	Id                int
	ExternalReference string
	OriginalReference string
	Scenario          string
	Notes             string
}

func openDatabaseConnection() (*sql.DB, error) {
	databaseConnection, err := sql.Open("mysql", "root:root@/duplicates")

	if err != nil {
		return nil, errors.New(fmt.Sprintf("couldn't connect to database %s", err.Error()))
	}
	databaseConnection.SetConnMaxLifetime(-1)
	databaseConnection.SetMaxOpenConns(10)
	databaseConnection.SetMaxIdleConns(10)

	return databaseConnection, nil
}

func insertImageIntoDatabaseSet(databaseConnection *sql.DB, databaseSetImage ForbiddenImageCreation) error {
	externalReference := databaseSetImage.externalReference
	siftDescriptor := databaseSetImage.siftDescriptor
	orbDescriptor := databaseSetImage.orbDescriptor
	briskDescriptor := databaseSetImage.briskDescriptor
	pHash := databaseSetImage.pHash

	_, err := databaseConnection.Exec(
		"INSERT INTO database_image (external_reference, sift_descriptor, orb_descriptor, brisk_descriptor , p_hash) VALUES (?, ?, ?, ?, ?)",
		externalReference,
		siftDescriptor,
		orbDescriptor,
		briskDescriptor,
		pHash,
	)

	if err != nil {
		return errors.New(fmt.Sprintf("couldn't insert %s into database %s", externalReference, err.Error()))
	}
	log.Println(fmt.Sprintf("Inserted %s into Database Set", externalReference))
	return nil
}

func retrieveFeatureImageChunk(
	databaseConnection *sql.DB,
	descriptorType string,
	offset int,
	limit int) (*[]FeatureImageEntity, error) {
	imageRows, err := databaseConnection.Query(
		fmt.Sprintf("SELECT external_reference, %s FROM database_image LIMIT ? OFFSET ?", descriptorType),
		limit,
		offset,
	)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("couldn't retreive database set images from database: %s", err.Error()))
	}
	defer imageRows.Close()

	var imageEntityChunk []FeatureImageEntity

	for imageRows.Next() {
		var image FeatureImageEntity

		var err = imageRows.Scan(
			&image.externalReference,
			&image.descriptors,
		)

		if err != nil {
			continue
		}

		imageEntityChunk = append(imageEntityChunk, image)

		//log.Println(fmt.Sprintf("Retrieved %s from Database Set", image.externalReference))
	}
	return &imageEntityChunk, nil
}

func retrievePHashImageChunk(
	databaseConnection *sql.DB,
	offset int,
	limit int) ([]PHashImageEntity, error) {
	imageRows, err := databaseConnection.Query(
		"SELECT external_reference, p_hash FROM database_image LIMIT ? OFFSET ?",
		limit,
		offset,
	)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("couldn't retreive database set images from database: %s", err.Error()))
	}
	defer imageRows.Close()

	var imageEntityChunk []PHashImageEntity

	for imageRows.Next() {
		var image PHashImageEntity

		var err = imageRows.Scan(
			&image.externalReference,
			&image.hash,
		)

		if err != nil {
			continue
		}

		imageEntityChunk = append(imageEntityChunk, image)

		//log.Println(fmt.Sprintf("Retrieved %s from Database Set", image.externalReference))
	}
	return imageEntityChunk, nil
}

func insertImageIntoSearchSet(databaseConnection *sql.DB, modifiedImage SearchImageCreation) error {
	externalReference := modifiedImage.externalReference
	originalReference := modifiedImage.originalReference
	scenario := modifiedImage.scenario
	notes := modifiedImage.modificationInfo

	_, err := databaseConnection.Exec(
		"INSERT INTO search_image (external_reference, original_reference, scenario, notes) VALUES (?, ?, ?, ?)",
		externalReference,
		originalReference,
		scenario,
		notes,
	)

	if err != nil {
		return errors.New(fmt.Sprintf("couldn't insert %s into search set %s", externalReference, err.Error()))
	}
	log.Println(fmt.Sprintf("Inserted %s into search set", externalReference))
	return nil
}

func retrieveChunkFromSearchSet(
	databaseConnection *sql.DB,
	scenario string,
	offset int,
	limit int) ([]SearchImageEntity, error) {
	imageRows, err := databaseConnection.Query(
		"SELECT * FROM search_image WHERE scenario = ? LIMIT ? OFFSET ?",
		scenario,
		limit,
		offset,
	)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("couldn't retreive search set images from database: %s", err.Error()))
	}
	defer imageRows.Close()

	var imageEntityChunk []SearchImageEntity

	for imageRows.Next() {
		var image SearchImageEntity

		err := imageRows.Scan(&image.Id, &image.ExternalReference, &image.OriginalReference, &image.Scenario,
			&image.Notes)

		if err != nil {
			continue
		}

		imageEntityChunk = append(imageEntityChunk, image)

		//log.Println(fmt.Sprintf("Retrieved %s from search set", image.ExternalReference))
	}
	return imageEntityChunk, nil
}
