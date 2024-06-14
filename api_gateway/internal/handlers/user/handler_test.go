package user

import (
	"testing"

	"github.com/julienschmidt/httprouter"
)

func TestRegister(t *testing.T) {
	router := httprouter.New()
	router.Lookup()

}
