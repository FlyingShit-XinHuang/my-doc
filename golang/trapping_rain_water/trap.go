package trap

// Trap returns the capacity of trapped rain water
func Trap(height []int) int {
	lMaxHeight, rMaxHeight := 0, 0
	lIdx, rIdx := 0, len(height)-1

	sum := 0
	findMaxOrSumUp := func(i int, max *int) {
		if height[i] > *max {
			*max = height[i]
		} else {
			// The accumulated capacity can be determined by the peak
			// because the other side is higher than it.
			sum += *max - height[i]
		}
	}

	// Traverse the slice from lower side to the higher.
	// Use two 'max' vars to record peaks during traverse.
	for lIdx < rIdx {
		if height[lIdx] < height[rIdx] {
			findMaxOrSumUp(lIdx, &lMaxHeight)
			lIdx++
		} else {
			findMaxOrSumUp(rIdx, &rMaxHeight)
			rIdx--
		}
	}

	return sum
}
