package auth

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestFunc(t *testing.T) {
	JWTone, err := MakeJWT(uuid.New(), "hehehehehehe", time.Duration(time.Hour.Seconds()))
	if err != nil {
		fmt.Println(err)
		return
	}
	JWTtwo, err := MakeJWT(uuid.New(), "hahahahahaha", time.Duration(time.Hour.Seconds()))
	if err != nil {
		fmt.Println(err)
		return
	}
	JWTthree, err := MakeJWT(uuid.New(), "harharharharhar", time.Duration(time.Hour.Seconds()))
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(JWTone)
	fmt.Println(JWTtwo)
	fmt.Println(JWTthree)
	fmt.Println("SUCCESS")
}
