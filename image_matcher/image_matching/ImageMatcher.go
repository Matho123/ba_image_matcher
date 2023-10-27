package image_matching

import (
	"gocv.io/x/gocv"
	"image_matcher/image_handling"
	"math/bits"
	"time"
)

const DistanceRatioThreshold = 0.8

const BFMatcher = "bfm"
const FlannMatcher = "flann"

var MatcherMapping = map[string]FeatureBasedImageMatcher{
	BFMatcher:    &BruteForceMatcher{gocv.NewBFMatcher()},
	FlannMatcher: &FLANNBasedMatcher{gocv.NewFlannBasedMatcher()},
}

type FeatureBasedImageMatcher interface {
	FindMatches(imageDescriptors1 *gocv.Mat, imageDescriptors2 *gocv.Mat) [][]gocv.DMatch
}

type BruteForceMatcher struct {
	matcher gocv.BFMatcher
}

func (bfm *BruteForceMatcher) FindMatches(imageDescriptors1 *gocv.Mat, imageDescriptors2 *gocv.Mat) [][]gocv.DMatch {
	return bfm.matcher.KnnMatch(*imageDescriptors1, *imageDescriptors2, 2)
}

type FLANNBasedMatcher struct {
	matcher gocv.FlannBasedMatcher
}

func (flann *FLANNBasedMatcher) FindMatches(imageDescriptors1 *gocv.Mat, imageDescriptors2 *gocv.Mat) [][]gocv.DMatch {
	k := 2

	//the amount of rows in a Descriptor Mat corresponds to the amount of descriptors/keypoints in an image.
	//If an image has less than two descriptor flann will throw an error for k = 2, when trying to build a k-tree.
	if imageDescriptors1.Rows() <= 1 || imageDescriptors2.Rows() <= 1 {
		k = 1
	}

	image_handling.ConvertImageDescriptorMat(imageDescriptors1, gocv.MatTypeCV32F)
	image_handling.ConvertImageDescriptorMat(imageDescriptors2, gocv.MatTypeCV32F)

	return flann.matcher.KnnMatch(*imageDescriptors1, *imageDescriptors2, k)
}

func FindHashMatchesPerThreshold(hash1 uint64, hash2 uint64, thresholds *[]int) (map[int]bool, time.Duration) {
	thresholdMap := make(map[int]bool)

	amountOfThresholds := len(*thresholds)
	var finalMatchingTime time.Duration

	for i := 0; i < amountOfThresholds; i++ {
		threshold := (*thresholds)[i]
		isMatch, matchingTime := HashesAreMatch(hash1, hash2, threshold, false)
		thresholdMap[threshold] = isMatch

		if i == amountOfThresholds-1 {
			finalMatchingTime = matchingTime
		}
	}
	return thresholdMap, finalMatchingTime
}

func HashesAreMatch(hash1 uint64, hash2 uint64, maxDistance int, debug bool) (bool, time.Duration) {
	matchingStart := time.Now()
	hammingDistance := calculateHammingDistance(hash1, hash2)
	matchingTime := time.Since(matchingStart)
	if debug {
		println("hamming distance: ", hammingDistance)
	}

	return hammingDistance <= maxDistance, matchingTime
}

func calculateHammingDistance(hash1, hash2 uint64) int {
	return bits.OnesCount64(hash1 ^ hash2)
}
