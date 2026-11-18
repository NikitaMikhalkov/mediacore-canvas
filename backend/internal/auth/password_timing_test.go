package auth

import (
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func TestMalformedSaltPerformsDummyPasswordWork(t *testing.T) {
	// The header and cost remain valid, but the first salt character is invalid.
	malformed := "$2a$10$!ipdF8SDXNLcv5X2e5KQjuHll6Gg6A3.U9l.zNE.4EwRrphPONhku"
	if _, err := bcrypt.Cost([]byte(malformed)); err != nil {
		t.Fatalf("fixture cost: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(malformed), []byte("wrong")); err == nil || errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		t.Fatalf("fixture did not produce a decoding error: %v", err)
	}
	// Use the fastest baseline sample to reduce sensitivity to scheduling delays.
	baseline := time.Duration(1<<63 - 1)
	for i := 0; i < 3; i++ {
		start := time.Now()
		verifyPassword("wrong", dummyPasswordHash)
		if elapsed := time.Since(start); elapsed < baseline {
			baseline = elapsed
		}
	}
	start := time.Now()
	if verifyPassword("wrong", malformed) {
		t.Fatal("malformed hash authenticated")
	}
	elapsed := time.Since(start)
	if elapsed < baseline/8 {
		t.Fatalf("malformed hash skipped dummy work: elapsed %v, dummy baseline %v", elapsed, baseline)
	}
}
