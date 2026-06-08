package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func GenerateOTP() (string, error) {
	const digits = "0123456789"
	const length = 6

	otp := make([]byte, length)
	for i := range otp {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", fmt.Errorf("failed to generate OTP digit: %w", err)
		}
		otp[i] = digits[n.Int64()]
	}

	return string(otp), nil
}
