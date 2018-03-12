package provide_test

import (
	"github.com/MatthewValentine/provide"
	"runtime"
	"strconv"
	"testing"
)

func assert(t *testing.T, condition bool, args ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		t.Fatal(append([]interface{}{file + ":" + strconv.Itoa(line) + ":"}, args...)...)
	}
}

func TestProvideBasic(t *testing.T) {
	p, err := provide.NewProvider(
		func() KrabbyPatty {
			return "jabberwocky"
		},
		func(spongebob Spongebob) InPineapple {
			return spongebob
		},
		func(patrick Patrick) UnderSea {
			return patrick
		},
	)
	assert(t, err == nil, err)

	var ip InPineapple
	var us UnderSea
	err = p.Provide(&ip, &us)
	assert(t, err == nil, err)
	assert(t, ip == Spongebob{Patty: "jabberwocky"}, ip)
	assert(t, us == Patrick{Patty: "jabberwocky"}, us)
}

type InPineapple interface {
	inPineapple()
}

type UnderSea interface {
	underSea()
}

type Spongebob struct {
	Patty KrabbyPatty `provide:""`
}

func (Spongebob) inPineapple() {}
func (Spongebob) underSea()    {}

type Patrick struct {
	Patty KrabbyPatty
}

func (Patrick) underSea() {}

func (star *Patrick) PleaseProvide(kp KrabbyPatty) {
	star.Patty = kp
}

type KrabbyPatty string
