package image_handling

import (
	"fmt"
	"gocv.io/x/gocv"
	"log"
)

const distanceRatioThreshold = 0.8

var ImageMatcherMapping = map[string]ImageMatcher{
	"bfm":   BruteForceMatcher{gocv.NewBFMatcher()},
	"flann": FLANNMatcher{gocv.NewFlannBasedMatcher()},
}

type ImageMatcher interface {
	FindMatches(imageDescriptors1 *gocv.Mat, imageDescriptors2 *gocv.Mat) [][]gocv.DMatch
	Close()
}

type BruteForceMatcher struct {
	matcher gocv.BFMatcher
}

func (bfm BruteForceMatcher) FindMatches(imageDescriptors1 *gocv.Mat, imageDescriptors2 *gocv.Mat) [][]gocv.DMatch {
	return bfm.matcher.KnnMatch(*imageDescriptors1, *imageDescriptors2, 2)
}

func (bfm BruteForceMatcher) Close() {
	bfm.matcher.Close()
}

type FLANNMatcher struct {
	matcher gocv.FlannBasedMatcher
}

func (flann FLANNMatcher) FindMatches(imageDescriptors1 *gocv.Mat, imageDescriptors2 *gocv.Mat) [][]gocv.DMatch {
	k := 2

	//the amount of rows in a Descriptor Mat corresponds to the amount of Descriptors/Keypoints in an image.
	//If an image has only one descriptor flann will throw an error for k = 2, when trying to build a k-tree.
	if imageDescriptors1.Size()[0] <= 1 || imageDescriptors2.Size()[0] <= 1 {
		k = 1
	}

	ConvertImageDescriptorMat(imageDescriptors1, gocv.MatTypeCV32F)
	ConvertImageDescriptorMat(imageDescriptors2, gocv.MatTypeCV32F)

	return flann.matcher.KnnMatch(*imageDescriptors1, *imageDescriptors2, k)
}

func (flann FLANNMatcher) Close() {
	flann.matcher.Close()
}

func DetermineSimilarity(matches [][]gocv.DMatch, similarityThreshold float64) (bool, *[]gocv.DMatch) {
	filteredMatches, maxDist := filterMatches(&matches)

	if len(*filteredMatches) == 0 {
		log.Println("no good matches found")
		return false, nil
	}

	averageNormalizedDistance := 0.0
	if maxDist > 0 {
		distanceSum := 0.0
		for _, match := range *filteredMatches {
			distanceSum += match.Distance
		}
		normalizedDistanceSum := distanceSum / maxDist
		averageNormalizedDistance = normalizedDistanceSum / float64(len(*filteredMatches))
	}
	similarityScore := 1.0 - averageNormalizedDistance

	println(fmt.Sprintf("Similarity score: %.2f", similarityScore))
	println("Amount of filtered matches:", len(*filteredMatches))
	return similarityScore >= similarityThreshold, filteredMatches
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

		if firstBestMatchDistance < distanceRatioThreshold*secondBestMatchDistance {
			filteredMatches = append(filteredMatches, firstBestMatch)

			if firstBestMatchDistance > maxDist {
				maxDist = firstBestMatchDistance
			}
		}
	}

	return &filteredMatches, maxDist
}

func HashesAreMatch(hash1 uint64, hash2 uint64, maxDistance int) bool {
	hammingDistance := calculateHammingDistance(hash1, hash2)
	log.Println("hamming distance: ", hammingDistance)
	return hammingDistance <= maxDistance
}
