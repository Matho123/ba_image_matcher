package service

import (
	"gocv.io/x/gocv"
	"log"
)

// TODO: needs to be tested
const MatchDistanceRatioThreshold = 0.7
const SimilarityThreshold = 0.8

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

func determineSimilarity(matches [][]gocv.DMatch) (bool, []gocv.DMatch) {
	var goodMatches []gocv.DMatch
	for _, matchPair := range matches {
		bestMatch := matchPair[0]
		secondBestMatch := matchPair[1]

		//TODO: Watch out for x/0
		ratio := bestMatch.Distance / secondBestMatch.Distance
		if ratio < MatchDistanceRatioThreshold {
			goodMatches = append(goodMatches, bestMatch)
		}
	}

	var distanceSum = 0.0
	for _, match := range goodMatches {
		distanceSum += match.Distance
	}

	//TODO: refine
	averageDistance := distanceSum / float64(len(goodMatches))
	similarityScore := 1.0 - averageDistance

	//log.Println("distanceSum: ", distanceSum)
	//log.Println("good matches: ", float64(len(goodMatches)))
	log.Println("similarity score: ", similarityScore)

	return similarityScore > SimilarityThreshold, goodMatches
}
