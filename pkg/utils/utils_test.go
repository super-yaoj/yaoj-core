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
	count  int
}

func (r *ratingRator) Rating() int {
	return r.rating
}
func (r *ratingRator) Rate(rating int) {
	r.rating = rating
}
func (r *ratingRator) Count() int {
	return r.count
}

func TestCalcRating(t *testing.T) {
	var a []*ratingRator
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
	for i := range a {
		a[i].count++
	}
	err = utils.CalcRating(a)
	if err != nil {
		t.Error(err)
		return
	}
	pp.Print(a)
}
