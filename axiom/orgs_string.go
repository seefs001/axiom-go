// Code generated by "stringer -type=Plan -linecomment -output=orgs_string.go"; DO NOT EDIT.

package axiom

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Free-1]
	_ = x[Trial-2]
	_ = x[Pro-3]
	_ = x[Enterprise-4]
}

const _Plan_name = "freetrialproenterprise"

var _Plan_index = [...]uint8{0, 4, 9, 12, 22}

func (i Plan) String() string {
	i -= 1
	if i >= Plan(len(_Plan_index)-1) {
		return "Plan(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _Plan_name[_Plan_index[i]:_Plan_index[i+1]]
}