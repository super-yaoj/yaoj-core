package utils_test

import (
	"bytes"
	"testing"

	"github.com/k0kubun/pp/v3"
	"github.com/super-yaoj/yaoj-core/pkg/utils"
)

func TestChecksum(t *testing.T) {
	sum := utils.ReaderChecksum(bytes.NewReader([]byte("hello1")))
	t.Log(sum)
}

type ratingRator struct {
	rating int
}

func (r *ratingRator) Rating() int {
	return r.rating
}
func (r *ratingRator) Rate(rating int) {
	r.rating = rating
}

func TestCalcRating(t *testing.T) {
	var a []utils.RatingRater
	a = append(a,
		&ratingRator{rating: 0},
		&ratingRator{rating: 100},
		&ratingRator{rating: 200},
		&ratingRator{rating: 300},
		&ratingRator{rating: 200},
		&ratingRator{rating: 100},
		&ratingRator{rating: 0},
	)
	err := utils.CalcRating(a)
	if err != nil {
		t.Error(err)
		return
	}
	utils.CalcRating(a)
	utils.CalcRating(a)
	err = utils.CalcRating(a)
	if err != nil {
		t.Error(err)
		return
	}
	pp.Print(a)
}
