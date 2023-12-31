package image_matching

import (
	"gocv.io/x/gocv"
	"image_matcher/image_analyzer"
	"image_matcher/image_database"
	"image_matcher/image_handling"
	"time"
)

const matchedHammingDistance = 12
const matchingPoolHammingDistance = 16
const similarityThreshold = 0.45

func HybridImageMatcher(
	orientedHashes []uint64, regularHash uint64, searchImageDescriptors *gocv.Mat, debug bool,
) (*[]string, int, time.Duration) {
	matchingPool, matchedImages, matchingTime := buildMatchingPool(orientedHashes, regularHash, debug)
	bfm := MatcherMapping[BFMatcher]

	totalMatchedImages := *matchedImages
	start := time.Now()
	for searchImageReference, descriptorBytes := range *matchingPool {
		originalImageDescriptors, _ :=
			image_handling.ConvertByteArrayToDescriptorMat(&descriptorBytes, image_analyzer.SIFT)
		matches := bfm.FindMatches(searchImageDescriptors, originalImageDescriptors)
		isMatch, _, _ := DetermineSimilarity(matches, similarityThreshold, false)
		if isMatch {
			totalMatchedImages = append(totalMatchedImages, searchImageReference)
		}
	}
	return &totalMatchedImages, len(*matchingPool), time.Since(start) + matchingTime
}

func buildMatchingPool(orientedHashes []uint64, regularHash uint64, debug bool) (*map[string][]byte, *[]string,
	time.Duration) {
	var matchedImages []string
	var totalMatchingTime time.Duration
	matchingPool := make(map[string][]byte)

	err := image_database.ApplyChunkedHybridRetrievalOperation(func(databaseImage image_database.HybridEntity) {
		if debug {
			println("Comparing to " + databaseImage.ExternalReference)
		}

		isMatch, hammingDistance, matchingTime :=
			HashesAreMatch(databaseImage.RegularHash, regularHash, matchingPoolHammingDistance, false)

		if !isMatch {
			orientedMatch, _, orientedHammingDistance, orientedMatchingTime :=
				MatchOrientedHashes(databaseImage.OrientedHash, orientedHashes, matchingPoolHammingDistance)
			isMatch, hammingDistance = orientedMatch, orientedHammingDistance
			matchingTime += orientedMatchingTime
		}

		if isMatch {
			if hammingDistance <= matchedHammingDistance {
				matchedImages = append(matchedImages, databaseImage.ExternalReference)
			} else {
				matchingPool[databaseImage.ExternalReference] = databaseImage.SiftDescriptors
			}
		}
		totalMatchingTime += matchingTime
	})
	if err != nil {
		return nil, nil, time.Duration(0)
	}

	return &matchingPool, &matchedImages, totalMatchingTime
}
