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
	//TODO: fix this
	bruteForceMatcher := gocv.NewBFMatcherWithParams(gocv.NormHamming, true)
	defer bruteForceMatcher.Close()

	convertedDescriptor1, convertedDescriptor2 := convertImageDescriptors(imageDescriptor1, imageDescriptor2)
	defer convertedDescriptor1.Close()
	defer convertedDescriptor2.Close()

	//TODO: Maybe sort matches by distance?
	return bruteForceMatcher.KnnMatch(convertedDescriptor1, convertedDescriptor2, 2)
}

type FLANNMatcher struct{}

func (FLANNMatcher) findMatches(imageDescriptor1 gocv.Mat, imageDescriptor2 gocv.Mat) [][]gocv.DMatch {
	flannBasedMatcher := gocv.NewFlannBasedMatcher()
	defer flannBasedMatcher.Close()

	//TODO: Maybe sort matches by distance?
	return flannBasedMatcher.KnnMatch(imageDescriptor1, imageDescriptor2, 2)
}

func convertImageDescriptors(descriptor1 gocv.Mat, descriptor2 gocv.Mat) (gocv.Mat, gocv.Mat) {
	convertedDescriptor1 := gocv.NewMat()
	defer convertedDescriptor1.Close()
	descriptor1.ConvertTo(&convertedDescriptor1, gocv.MatTypeCV32F)

	convertedDescriptor2 := gocv.NewMat()
	defer convertedDescriptor2.Close()
	descriptor2.ConvertTo(&convertedDescriptor2, gocv.MatTypeCV32F)

	return convertedDescriptor1, convertedDescriptor2
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

	log.Println("distanceSum: ", distanceSum)
	log.Println("good matches: ", float64(len(goodMatches)))
	log.Println("similarity score: ", similarityScore)
	return similarityScore > SimilarityThreshold, goodMatches
}
