package image_matching

import (
	"fmt"
	"gocv.io/x/gocv"
	"log"
)

func FindDescriptorMatchesPerThreshold(
	matches [][]gocv.DMatch, matchedPerThreshold *map[float64][]string, originalReference string,
) {
	for threshold, matchedImages := range *matchedPerThreshold {
		isMatch, _, _ := DetermineSimilarity(matches, threshold, false)
		if isMatch {
			matchedImages = append(matchedImages, originalReference)
			(*matchedPerThreshold)[threshold] = matchedImages
		}
	}
}

func DetermineSimilarity(matches [][]gocv.DMatch, similarityThreshold float64, debug bool) (
	bool,
	float64,
	*[]gocv.DMatch,
) {
	filteredMatches, maxDist := filterMatches(&matches)

	if len(*filteredMatches) == 0 {
		if debug {
			log.Println("no good matches found")
		}

		return false, 0.0, nil
	}

	filteredMatchRatio := float64(len(*filteredMatches)) / float64(len(matches))

	averageNormalizedDistance := 0.0
	if maxDist > 0 {
		distanceSum := 0.0
		for _, match := range *filteredMatches {
			distanceSum += match.Distance
		}
		normalizedDistanceSum := distanceSum / maxDist
		averageNormalizedDistance = normalizedDistanceSum / float64(len(*filteredMatches))
	}
	similarityScore := 0.5*(1.0-averageNormalizedDistance) + 0.5*filteredMatchRatio

	if debug {
		println(fmt.Sprintf("Similarity score: %.2f", similarityScore))
		println(fmt.Sprintf("Average match distance: %.2f", averageNormalizedDistance))
		//println("Amount of matches:", len(matches))
		//println("Amount of filtered matches:", len(*filteredMatches))
		println(fmt.Sprintf("Filtered to unfiltered match ratio: %.2f", filteredMatchRatio))
	}

	return similarityScore >= similarityThreshold, similarityScore, filteredMatches
}

// applying ratio test according to D. Lowe
func filterMatches(matches *[][]gocv.DMatch) (*[]gocv.DMatch, float64) {
	var filteredMatches []gocv.DMatch
	var maxDist float64

	for _, matchPair := range *matches {
		if len(matchPair) < 2 {
			filteredMatches = append(filteredMatches, matchPair[0])
			continue
		}

		firstBestMatch := matchPair[0]
		secondBestMatch := matchPair[1]

		firstBestMatchDistance := firstBestMatch.Distance
		secondBestMatchDistance := secondBestMatch.Distance

		if firstBestMatchDistance < DistanceRatioThreshold*secondBestMatchDistance {
			filteredMatches = append(filteredMatches, firstBestMatch)

			if firstBestMatchDistance > maxDist {
				maxDist = firstBestMatchDistance
			}
		}
	}

	return &filteredMatches, maxDist
}
