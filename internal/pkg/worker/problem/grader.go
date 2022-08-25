package problemruntime

import "github.com/super-yaoj/yaoj-core/pkg/problem"

type Grader struct {
	method    problem.CalcMethod
	fullscore float64
	ntask     int
	current   float64
}

func NewGrader(method problem.CalcMethod, fullscore float64, ntask int) *Grader {
	res := &Grader{
		method:    method,
		fullscore: fullscore,
		ntask:     ntask,
	}
	if method == problem.Mmin {
		res.current = fullscore
	}
	return res
}

func (r *Grader) TaskFullscore() float64 {
	if r.method == problem.Msum {
		return r.fullscore / float64(r.ntask)
	} else {
		return r.fullscore
	}
}

func (r *Grader) Add(score float64) {
	if r.method == problem.Msum {
		r.current += score
	} else if r.method == problem.Mmin {
		if r.current > score {
			r.current = score
		}
	} else {
		if r.current < score {
			r.current = score
		}
	}
}

func (r *Grader) Skipable() bool {
	if r.method == problem.Mmin && r.current == 0 {
		return true
	}
	if r.method == problem.Mmax && r.current == r.fullscore {
		return true
	}
	return false
}

func (r *Grader) Sum() float64 {
	return r.current
}
