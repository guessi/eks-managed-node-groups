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

func ValidateNodegroupSize(desiredSize, minSize, maxSize int32) bool {
	switch {
	case desiredSize < 0:
		fmt.Println("Requested desired capacity can't be less than 0.")
		return false
	case minSize < 0:
		fmt.Println("Requested min capacity can't be less than 0.")
		return false
	case maxSize < 1:
		fmt.Println("Requested max capacity can't be less than 1.")
		return false
	case minSize > desiredSize:
		fmt.Printf("Minimum capacity %d can't be greater than desired size %d.\n", minSize, desiredSize)
		return false
	case minSize > maxSize:
		fmt.Printf("Minimum capacity %d can't be greater than max size %d.\n", minSize, maxSize)
		return false
	}
	return true
}
