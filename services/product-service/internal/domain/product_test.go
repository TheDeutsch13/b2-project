package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrimaryVariant(t *testing.T) {
	assert.Equal(t, "pro", PrimaryVariant([]string{"  Pro  ", ""}))
	assert.Equal(t, "стандарт", PrimaryVariant([]string{"", "  "}))
	assert.Equal(t, "стандарт", PrimaryVariant(nil))
}

func TestCalcRatingFromReviews(t *testing.T) {
	avg, count := CalcRatingFromReviews(nil)
	assert.Equal(t, 0, count)
	assert.Equal(t, float64(0), avg)

	avg, count = CalcRatingFromReviews([]ProductReview{
		{Rating: 5},
		{Rating: 3},
		{Rating: 0},
		{Rating: 6},
	})
	assert.Equal(t, 2, count)
	assert.InDelta(t, 4.0, avg, 0.01)
}
