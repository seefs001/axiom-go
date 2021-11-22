// Code generated by "stringer -type=AggregationOp -linecomment -output=aggregation_string.go"; DO NOT EDIT.

package query

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[emptyAggregationOp-0]
	_ = x[OpCount-1]
	_ = x[OpCountDistinct-2]
	_ = x[OpSum-3]
	_ = x[OpAvg-4]
	_ = x[OpMin-5]
	_ = x[OpMax-6]
	_ = x[OpTopk-7]
	_ = x[OpPercentiles-8]
	_ = x[OpHistogram-9]
	_ = x[OpVariance-10]
	_ = x[OpStandardDeviation-11]
	_ = x[OpCountIf-12]
	_ = x[OpCountDistinctIf-13]
}

const _AggregationOp_name = "countdistinctsumavgminmaxtopkpercentileshistogramvariancestdevcountifdistinctif"

var _AggregationOp_index = [...]uint8{0, 0, 5, 13, 16, 19, 22, 25, 29, 40, 49, 57, 62, 69, 79}

func (i AggregationOp) String() string {
	if i >= AggregationOp(len(_AggregationOp_index)-1) {
		return "AggregationOp(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _AggregationOp_name[_AggregationOp_index[i]:_AggregationOp_index[i+1]]
}