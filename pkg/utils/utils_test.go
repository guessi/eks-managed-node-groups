package utils

import "testing"

func TestValidateNodegroupSize(t *testing.T) {
	tests := []struct {
		name        string
		desiredSize int32
		minSize     int32
		maxSize     int32
		wantErr     bool
	}{
		{"accepts desired within min max range", 1, 0, 5, false},
		{"accepts all sizes equal", 3, 3, 3, false},
		{"accepts desired equals max", 5, 0, 5, false},
		{"rejects negative desired", -1, 0, 5, true},
		{"rejects negative min", 1, -1, 5, true},
		{"rejects max less than 1", 0, 0, 0, true},
		{"rejects min greater than desired", 1, 2, 5, true},
		{"rejects min greater than max", 6, 6, 5, true},
		{"rejects desired greater than max", 10, 0, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNodegroupSize(tt.desiredSize, tt.minSize, tt.maxSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNodegroupSize(%d, %d, %d) error = %v, wantErr %v",
					tt.desiredSize, tt.minSize, tt.maxSize, err, tt.wantErr)
			}
		})
	}
}
