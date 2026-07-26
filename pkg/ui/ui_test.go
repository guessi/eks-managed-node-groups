package ui

import "testing"

func i32(v int32) *int32 { return &v }

func TestOptionsValidate(t *testing.T) {
	tests := []struct {
		name    string
		opts    Options
		wantErr bool
	}{
		{"accepts empty options", Options{}, false},
		{"accepts valid region", Options{Region: "ap-northeast-1"}, false},
		{"rejects region with invalid characters", Options{Region: "us-east-1.evil.com/"}, true},
		{"rejects region with uppercase", Options{Region: "US-EAST-1"}, true},
		{"accepts managed type", Options{NodeGroupType: "managed"}, false},
		{"accepts self-managed type", Options{NodeGroupType: "self-managed"}, false},
		{"rejects unknown type", Options{NodeGroupType: "foo"}, true},
		{"accepts text output", Options{Output: "text"}, false},
		{"accepts json output", Options{Output: "json"}, false},
		{"rejects unknown output", Options{Output: "yaml"}, true},
		{"accepts all three sizes", Options{DesiredSize: i32(1), MinSize: i32(0), MaxSize: i32(2)}, false},
		{"rejects desired only", Options{DesiredSize: i32(1)}, true},
		{"rejects min only", Options{MinSize: i32(0)}, true},
		{"rejects desired and min without max", Options{DesiredSize: i32(1), MinSize: i32(0)}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResolveNodegroupName(t *testing.T) {
	nodegroups := []string{"ng-1", "ng-2"}

	t.Run("accepts existing nodegroup", func(t *testing.T) {
		name, err := resolveNodegroupName(Options{NodegroupName: "ng-2"}, nodegroups)
		if err != nil || name != "ng-2" {
			t.Errorf("got (%q, %v), want (%q, nil)", name, err, "ng-2")
		}
	})

	t.Run("rejects unknown nodegroup", func(t *testing.T) {
		_, err := resolveNodegroupName(Options{NodegroupName: "missing", ClusterName: "c"}, nodegroups)
		if err == nil {
			t.Error("expected error for unknown nodegroup, got nil")
		}
	})
}

func TestResolveNodegroupSize(t *testing.T) {
	t.Run("accepts valid sizes from flags", func(t *testing.T) {
		desired, minSize, maxSize, err := resolveNodegroupSize(Options{DesiredSize: i32(1), MinSize: i32(0), MaxSize: i32(2)})
		if err != nil || desired != 1 || minSize != 0 || maxSize != 2 {
			t.Errorf("got (%d, %d, %d, %v), want (1, 0, 2, nil)", desired, minSize, maxSize, err)
		}
	})

	t.Run("rejects desired greater than max from flags", func(t *testing.T) {
		_, _, _, err := resolveNodegroupSize(Options{DesiredSize: i32(10), MinSize: i32(0), MaxSize: i32(5)})
		if err == nil {
			t.Error("expected error for desired greater than max, got nil")
		}
	})
}

func TestScalingConfigDrifted(t *testing.T) {
	tests := []struct {
		name                                string
		latestDesired, latestMin, latestMax *int32
		want                                bool
	}{
		{"no drift when values match", i32(1), i32(0), i32(5), false},
		{"drift on desired change", i32(2), i32(0), i32(5), true},
		{"drift on min change", i32(1), i32(1), i32(5), true},
		{"drift on max change", i32(1), i32(0), i32(4), true},
		{"drift on nil desired", nil, i32(0), i32(5), true},
		{"drift on nil min", i32(1), nil, i32(5), true},
		{"drift on nil max", i32(1), i32(0), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scalingConfigDrifted(1, 0, 5, tt.latestDesired, tt.latestMin, tt.latestMax)
			if got != tt.want {
				t.Errorf("scalingConfigDrifted() = %v, want %v", got, tt.want)
			}
		})
	}
}
