package service

import (
	"gocv.io/x/gocv"
	"log"
)

// TODO: needs to be tested
const SimilarityThreshold = 0.7

var imageMatcherMapping = map[string]ImageMatcher{
	"bfm":   BruteForceMatcher{},
	"flann": FLANNMatcher{},
}

type ImageMatcher interface {
	findMatches(imageDescriptor1 gocv.Mat, imageDescriptor2 gocv.Mat) [][]gocv.DMatch
}

type BruteForceMatcher struct{}

func (BruteForceMatcher) findMatches(imageDescriptor1 gocv.Mat, imageDescriptor2 gocv.Mat) [][]gocv.DMatch {
	bruteForceMatcher := gocv.NewBFMatcher()
	defer bruteForceMatcher.Close()

	convertImageDescriptors(&imageDescriptor1, &imageDescriptor2, gocv.MatTypeCV32F)

	return bruteForceMatcher.KnnMatch(imageDescriptor1, imageDescriptor2, 2)
}

type FLANNMatcher struct{}

func (FLANNMatcher) findMatches(imageDescriptor1 gocv.Mat, imageDescriptor2 gocv.Mat) [][]gocv.DMatch {
	flannBasedMatcher := gocv.NewFlannBasedMatcher()
	defer flannBasedMatcher.Close()

	convertImageDescriptors(&imageDescriptor1, &imageDescriptor2, gocv.MatTypeCV32F)

	return flannBasedMatcher.KnnMatch(imageDescriptor1, imageDescriptor2, 2)
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
func determineSimilarity(matches [][]gocv.DMatch) (bool, []gocv.DMatch) {
	var filteredMatches []gocv.DMatch
	var maxDist = 0.0
	var minDist = 0.0

	for _, matchPair := range matches {
		firstBestMatch := matchPair[0]
		secondBestMatch := matchPair[1]

		firstBestMatchDistance := firstBestMatch.Distance
		secondBestMatchDistance := secondBestMatch.Distance

		if firstBestMatchDistance < 0.75*secondBestMatchDistance {
			filteredMatches = append(filteredMatches, firstBestMatch)

			if firstBestMatchDistance > maxDist {
				maxDist = firstBestMatchDistance
			}
			if firstBestMatchDistance < minDist || minDist == 0 {
				minDist = firstBestMatchDistance
			}
		}
	}

	var normalizedDistanceSum = 0.0
	for _, match := range filteredMatches {
		if maxDist == 0 {
			normalizedDistanceSum += 0
		} else {
			normalizedDistanceSum += (match.Distance - minDist) / (maxDist - minDist)
		}
	}

	if len(filteredMatches) == 0 {
		return false, nil
	}
	averageNormalizedDistance := normalizedDistanceSum / float64(len(filteredMatches))
	similarityScore := 1.0 - averageNormalizedDistance

	log.Println("similarity score: ", similarityScore)

	return similarityScore > SimilarityThreshold, filteredMatches
}
