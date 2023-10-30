package image_database

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

type ForbiddenImageCreation struct {
	ExternalReference     string
	SiftDescriptor        []byte
	OrbDescriptor         []byte
	BriskDescriptor       []byte
	PHash                 uint64
	RotationInvariantHash uint64
}

type SearchImageCreation struct {
	ExternalReference string
	OriginalReference string
	Scenario          string
	ModificationInfo  string
}

type FeatureImageEntity struct {
	ExternalReference string
	Descriptors       []byte
}

type PHashImageEntity struct {
	ExternalReference string
	Hash              uint64
}

type SearchImageEntity struct {
	Id                int
	ExternalReference string
	OriginalReference string
	Scenario          string
	Notes             string
}

type HybridEntity struct {
	ExternalReference string
	OrientedHash      uint64
	RegularHash       uint64
	SiftDescriptors   []byte
}

func openDatabaseConnection() (*sql.DB, error) {
	databaseConnection, err := sql.Open("mysql", "root:root@/duplicates")

	if err != nil {
		log.Fatal("Couldn't connect to database!", err)
		return nil, errors.New(fmt.Sprintf("couldn't connect to database %s", err.Error()))
	}
	databaseConnection.SetConnMaxLifetime(-1)
	databaseConnection.SetMaxOpenConns(10)
	databaseConnection.SetMaxIdleConns(10)

	return databaseConnection, nil
}

func GetForbiddenReferences(databaseConnection *sql.DB) (*[]string, error) {
	imageRows, err := databaseConnection.Query("SELECT external_reference FROM forbidden_image")
	if err != nil {
		return nil, errors.New(fmt.Sprintf("couldn't retrieve external references from database %s", err.Error()))
	}
	defer imageRows.Close()

	var forbiddenReferences []string

	for imageRows.Next() {
		var externalReference string

		var err = imageRows.Scan(&externalReference)

		if err != nil {
			continue
		}

		forbiddenReferences = append(forbiddenReferences, externalReference)
	}

	return &forbiddenReferences, nil
}

func InsertRotationHashIntoDatabase(databaseConnection *sql.DB, externalReference string, rotHash uint64) error {
	_, err := databaseConnection.Exec(
		"UPDATE forbidden_image SET rotation_phash = ? WHERE external_reference = ?",
		rotHash,
		externalReference,
	)
	if err != nil {
		return errors.New(fmt.Sprintf("couldn't update %s into database %s", externalReference, err.Error()))
	}
	log.Println(fmt.Sprintf("Updated %s into Database Set with %d", externalReference, rotHash))
	return nil
}

func InsertImageIntoDatabaseSet(databaseConnection *sql.DB, databaseSetImage ForbiddenImageCreation) error {
	externalReference := databaseSetImage.ExternalReference
	siftDescriptor := databaseSetImage.SiftDescriptor
	orbDescriptor := databaseSetImage.OrbDescriptor
	briskDescriptor := databaseSetImage.BriskDescriptor
	pHash := databaseSetImage.PHash
	rotationInvariantHash := databaseSetImage.RotationInvariantHash

	_, err := databaseConnection.Exec(
		"INSERT INTO forbidden_image (external_reference, sift_descriptor, orb_descriptor, brisk_descriptor, p_hash, rotation_phash) VALUES (?, ?, ?, ?, ?, ?)",
		externalReference,
		siftDescriptor,
		orbDescriptor,
		briskDescriptor,
		pHash,
		rotationInvariantHash,
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
		fmt.Sprintf("SELECT external_reference, %s FROM forbidden_image LIMIT ? OFFSET ?", descriptorType),
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
			&image.ExternalReference,
			&image.Descriptors,
		)

		if err != nil {
			continue
		}

		imageEntityChunk = append(imageEntityChunk, image)

		//log.Println(fmt.Sprintf("Retrieved %s from Database Set", image.ExternalReference))
	}
	return &imageEntityChunk, nil
}

func retrievePHashImageChunk(databaseConnection *sql.DB, offset int, limit int) (*[]PHashImageEntity, error) {
	imageRows, err := databaseConnection.Query(
		"SELECT external_reference, p_hash FROM forbidden_image LIMIT ? OFFSET ?",
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
			&image.ExternalReference,
			&image.Hash,
		)

		if err != nil {
			continue
		}

		imageEntityChunk = append(imageEntityChunk, image)

		//log.Println(fmt.Sprintf("Retrieved %s from Database Set", image.ExternalReference))
	}
	return &imageEntityChunk, nil
}

func retrieveHybridChunk(databaseConnection *sql.DB, offset int, limit int) (*[]HybridEntity, error) {
	imageRows, err := databaseConnection.Query(
		"SELECT external_reference, sift_descriptor, rotation_phash, p_hash FROM forbidden_image LIMIT ? OFFSET ?",
		limit,
		offset,
	)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("couldn't retreive database set images from database: %s", err.Error()))
	}
	defer imageRows.Close()

	var imageEntityChunk []HybridEntity

	for imageRows.Next() {
		var image HybridEntity

		var err = imageRows.Scan(
			&image.ExternalReference,
			&image.SiftDescriptors,
			&image.OrientedHash,
			&image.RegularHash,
		)

		if err != nil {
			continue
		}

		imageEntityChunk = append(imageEntityChunk, image)

		//log.Println(fmt.Sprintf("Retrieved %s from Database Set", image.ExternalReference))
	}
	return &imageEntityChunk, nil
}

func InsertImageIntoSearchSet(databaseConnection *sql.DB, modifiedImage SearchImageCreation) error {
	externalReference := modifiedImage.ExternalReference
	originalReference := modifiedImage.OriginalReference
	scenario := modifiedImage.Scenario
	notes := modifiedImage.ModificationInfo

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

func RetrieveChunkFromSearchSet(
	databaseConnection *sql.DB,
	scenario string,
	offset int,
	limit int,
) ([]SearchImageEntity, error) {
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
