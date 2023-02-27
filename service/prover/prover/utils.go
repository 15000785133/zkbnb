package prover

import (
	"strconv"
)

func ConvertIntArrToStringArr(nums []int) []string {
	strArr := make([]string, len(nums))
	for i, num := range nums {
		strArr[i] = strconv.Itoa(num)
	}
	return strArr
}
