package utils

import (
	"errors"
	"fmt"
	"log"
	"strconv"
)

func IsInteger(input string) error {
	_, err := strconv.Atoi(input)
	if err != nil {
		return errors.New("invalid input")
	}
	return nil
}

func ParseInt32(input string) int32 {
	if input == "" {
		return *new(int32)
	}

	value, err := strconv.ParseInt(input, 10, 32)
	if err != nil {
		log.Fatal(err)
	}

	return int32(value)
}

func ValidateNodegroupSize(desiredSize, minSize, maxSize int32) error {
	switch {
	case desiredSize < 0:
		return errors.New("requested desired capacity can't be less than 0")
	case minSize < 0:
		return errors.New("requested min capacity can't be less than 0")
	case maxSize < 1:
		return errors.New("requested max capacity can't be less than 1")
	case minSize > desiredSize:
		return fmt.Errorf("minimum capacity %d can't be greater than desired size %d", minSize, desiredSize)
	case minSize > maxSize:
		return fmt.Errorf("minimum capacity %d can't be greater than max size %d", minSize, maxSize)
	}
	return nil
}
