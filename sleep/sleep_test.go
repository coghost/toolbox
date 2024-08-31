package sleep

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SleepSuite struct {
	suite.Suite
}

func TestSleep(t *testing.T) {
	suite.Run(t, new(SleepSuite))
}

func (s *SleepSuite) TestCeilInt() {
	assert.Equal(s.T(), 2, CeilInt(3, 2))
	assert.Equal(s.T(), 2, CeilInt(4, 2))
	assert.Equal(s.T(), 3, CeilInt(5, 2))
}

func (s *SleepSuite) TestRandFloatX1k() {
	for i := 0; i < 100; i++ {
		result := RandFloatX1k(1.0, 2.0)
		assert.GreaterOrEqual(s.T(), result, 1000)
		assert.LessOrEqual(s.T(), result, 2000)
	}
}

func (s *SleepSuite) TestRandRange() {
	start := time.Now()
	slept := RandRange(1.0, 2.0)
	duration := time.Since(start)

	assert.GreaterOrEqual(s.T(), slept, 1000)
	assert.LessOrEqual(s.T(), slept, 2000)
	assert.InDelta(s.T(), float64(slept), duration.Milliseconds(), 10)
}

func (s *SleepSuite) TestRandN() {
	start := time.Now()
	slept := RandN(3.0)
	duration := time.Since(start)

	assert.InDelta(s.T(), 3000, slept, 1500)
	assert.InDelta(s.T(), 3000, duration.Milliseconds(), 1500)
}

func (s *SleepSuite) TestPTFunctions() {
	testCases := []struct {
		name     string
		function func() int
		expected int
	}{
		{"PT5s", PT5s, 5000},
		{"PT1s", PT1s, 1000},
		{"PT2s", PT2s, 2000},
		{"PT3s", PT3s, 3000},
		{"PT4s", PT4s, 4000},
		{"PT1Ms", PT1Ms, 1},
		{"PT10Ms", PT10Ms, 10},
		{"PT100Ms", PT100Ms, 100},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			start := time.Now()
			slept := tc.function()
			duration := time.Since(start)

			assert.InDelta(s.T(), tc.expected, slept, float64(tc.expected)/2)
			assert.InDelta(s.T(), tc.expected, duration.Milliseconds(), float64(tc.expected)/2)
		})
	}
}

func (s *SleepSuite) TestRandNS() {
	testCases := []struct {
		num      float64
		scale    float64
		expected int
	}{
		{1.0, 2.0, 1000},
		{2.0, 1.5, 2000},
		{0.5, 3.0, 500},
	}

	for _, tc := range testCases {
		start := time.Now()
		slept := randNS(tc.num, tc.scale)
		duration := time.Since(start)

		minExpected := int(tc.num / tc.scale * 1000)
		maxExpected := int((tc.num + tc.num - tc.num/tc.scale) * 1000)

		assert.GreaterOrEqual(s.T(), slept, minExpected)
		assert.LessOrEqual(s.T(), slept, maxExpected)
		assert.InDelta(s.T(), float64(slept), duration.Milliseconds(), 10)
	}
}
