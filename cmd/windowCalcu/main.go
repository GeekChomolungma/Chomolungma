package main

import "fmt"

func main() {
	tw, dl := timeWindowAtEndTime(1627747200, 1629457200)
	str := fmt.Sprintf("tw is %v, datalength is %d", tw, dl)
	fmt.Println(str)
}

func timeWindowAtEndTime(startTime, endTime int64) ([][]int64, int) {
	var divisor int64 = 300
	var timeWindow [][]int64

	dataLength := ((endTime - startTime) / divisor) + 1
	windowLength := dataLength / 300
	for i := 0; int64(i) < windowLength; i++ {
		fmt.Println("i is ", i)
		start := startTime + int64(i)*300*divisor           // 0                  60*300
		end := startTime + int64(i+1)*300*divisor - divisor // 60*300             60*300 + 60*300
		timeElement := []int64{start, end}                  // [0:60:60*300)      [60*300:60:60*300*2)
		timeWindow = append(timeWindow, timeElement)        // [) [) [) [) [)
	}

	residual := dataLength % 300
	if residual > 0 {
		fmt.Println("windowLength is ", windowLength)
		startResidual := startTime + windowLength*300*divisor
		timeElementResidual := []int64{startResidual, endTime}
		timeWindow = append(timeWindow, timeElementResidual)
	}
	return timeWindow, int(dataLength)
}
