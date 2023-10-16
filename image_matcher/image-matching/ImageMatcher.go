package image_matching

import (
	"gocv.io/x/gocv"
	"log"
)

const DistanceRatioThreshold = 0.8

var ImageMatcherMapping = map[string]ImageMatcher{
	"bfm":   BruteForceMatcher{gocv.NewBFMatcher()},
	"flann": FLANNMatcher{gocv.NewFlannBasedMatcher()},
}

type ImageMatcher interface {
	FindMatches(imageDescriptor1 gocv.Mat, imageDescriptor2 gocv.Mat) [][]gocv.DMatch
	Close()
}

type BruteForceMatcher struct {
	matcher gocv.BFMatcher
}

func (bfm BruteForceMatcher) FindMatches(imageDescriptor1 gocv.Mat, imageDescriptor2 gocv.Mat) [][]gocv.DMatch {
	convertImageDescriptors(&imageDescriptor1, &imageDescriptor2, gocv.MatTypeCV32F)

	return bfm.matcher.KnnMatch(imageDescriptor1, imageDescriptor2, 2)
}

func (bfm BruteForceMatcher) Close() {
	bfm.matcher.Close()
}

type FLANNMatcher struct {
	matcher gocv.FlannBasedMatcher
}

func (flann FLANNMatcher) FindMatches(imageDescriptor1 gocv.Mat, imageDescriptor2 gocv.Mat) [][]gocv.DMatch {
	convertImageDescriptors(&imageDescriptor1, &imageDescriptor2, gocv.MatTypeCV32F)

	return flann.matcher.KnnMatch(imageDescriptor1, imageDescriptor2, 2)
}

func (flann FLANNMatcher) Close() {
	flann.matcher.Close()
}

func convertImageDescriptors(descriptor1 *gocv.Mat, descriptor2 *gocv.Mat, goalType gocv.MatType) (*gocv.Mat, *gocv.Mat) {
	if descriptor1.Type() != goalType {
		descriptor1.ConvertTo(descriptor1, gocv.MatTypeCV32F)
	}
	if descriptor2.Type() != goalType {
		descriptor2.ConvertTo(descriptor2, gocv.MatTypeCV32F)
	}
	return descriptor1, descriptor2
}

// TODO: refine with testing
func DetermineSimilarity(matches [][]gocv.DMatch, similarityThreshold float64) (bool, []gocv.DMatch) {
	var filteredMatches []gocv.DMatch
	var maxDist = 0.0

	//ratio test according to D. Lowe
	for _, matchPair := range matches {
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

	if len(filteredMatches) == 0 {
		log.Println("no good matches found")
		return false, nil
	}

	//similarity score calculation
	averageNormalizedDistance := 0.0
	if maxDist > 0 {
		distanceSum := 0.0
		for _, match := range filteredMatches {
			println(match.Distance)
			distanceSum += match.Distance
		}
		normalizedDistanceSum := distanceSum / maxDist
		averageNormalizedDistance = normalizedDistanceSum / float64(len(filteredMatches))
	}
	similarityScore := 1.0 - averageNormalizedDistance

	log.Println("similarity score: ", similarityScore)
	log.Println(len(matches), len(filteredMatches))
	return similarityScore >= similarityThreshold, filteredMatches
}

func HashesAreMatch(hash1 uint64, hash2 uint64, maxDistance int) bool {
	hammingDistance := calculateHammingDistance(hash1, hash2)
	log.Println("hamming distance: ", hammingDistance)
	return hammingDistance <= maxDistance
}
