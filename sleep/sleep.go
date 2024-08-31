//nolint:mnd
package sleep

import (
	"math"
	"math/rand"
	"time"
)

// CeilInt returns the ceiling of a/b as an integer.
func CeilInt(a, b int) int {
	return int(math.Ceil(float64(a) / float64(b)))
}

// RandFloatX1k returns a random integer value between (min * 1000) and (max * 1000).
func RandFloatX1k(minNum, maxNum float64) int {
	if minNum == maxNum {
		return int(minNum)
	}

	secToMs := 1000.0

	minI, maxI := int(math.Round(minNum*secToMs)), int(math.Round(maxNum*secToMs))
	x1k := minI + rand.Intn(maxI-minI)

	return x1k
}

// RandRange sleeps for a random duration between minNum and maxNum seconds.
//
// @return actual sleep duration in milliseconds
func RandRange(minNum, maxNum float64, msg ...string) int {
	slept := RandFloatX1k(minNum, maxNum)
	time.Sleep(time.Duration(slept) * time.Millisecond)

	return slept
}

func randNS(num float64, scales ...float64) int {
	scale := 2.0
	if len(scales) > 0 {
		scale = scales[0]
	}

	minNum := num / scale
	maxNum := num + num - minNum

	return RandRange(minNum, maxNum)
}

// RandN sleeps for a random duration around n seconds.
// @param n: target sleep duration in seconds
// @return actual sleep duration in milliseconds
func RandN(n float64) int {
	return randNS(n)
}

// PT5s sleeps for a random duration around 5 seconds.
// @return actual sleep duration in milliseconds
func PT5s() int {
	return randNS(5)
}

func PT1s() int {
	return randNS(1)
}

func PT2s() int {
	return randNS(2)
}

func PT3s() int {
	return randNS(3)
}

func PT4s() int {
	return randNS(4)
}

func PT1Ms() int {
	return randNS(0.001)
}

// PT10Ms sleeps on avg 10 ms
func PT10Ms() int {
	return randNS(0.01)
}

// PT100Ms sleeps on avg 100 ms
func PT100Ms() int {
	return randNS(0.1)
}
