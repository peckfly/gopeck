package rand

import (
	"bytes"
	"crypto/rand"
	"errors"
)

// define a flag that generates a random string
const (
	Digit = 1 << iota
	LowerCase
	UpperCase
)

var (
	digits           = []byte("0123456789")
	lowerCaseLetters = []byte("abcdefghijklmnopqrstuvwxyz")
	upperCaseLetters = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

// ErrInvalidFlag definition error
var (
	ErrInvalidFlag = errors.New("invalid flag")
)

// Random generate a random string specifying the length of the random number
// and the random flag
func Random(length, flag int) (string, error) {
	if length < 1 {
		length = 6
	}

	source, err := getFlagSource(flag)
	if err != nil {
		return "", err
	}

	b, err := randomBytesMod(length, byte(len(source)))
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	for _, c := range b {
		buf.WriteByte(source[c])
	}

	return buf.String(), nil
}

func getFlagSource(flag int) ([]byte, error) {
	var source []byte

	if flag&Digit > 0 {
		source = append(source, digits...)
	}

	if flag&LowerCase > 0 {
		source = append(source, lowerCaseLetters...)
	}

	if flag&UpperCase > 0 {
		source = append(source, upperCaseLetters...)
	}

	sourceLen := len(source)
	if sourceLen == 0 {
		return nil, ErrInvalidFlag
	}
	return source, nil
}

func randomBytesMod(length int, mod byte) ([]byte, error) {
	b := make([]byte, length)
	max := 255 - 255%mod
	i := 0

ROOT:
	for {
		r, err := randomBytes(length + length/4)
		if err != nil {
			return nil, err
		}

		for _, c := range r {
			if c >= max {
				// Skip this number to avoid modulo bias
				continue
			}

			b[i] = c % mod
			i++
			if i == length {
				break ROOT
			}
		}

	}

	return b, nil
}

func randomBytes(length int) ([]byte, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}
