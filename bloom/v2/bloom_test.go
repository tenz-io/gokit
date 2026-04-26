package bloom

import "testing"

func Test_optimalParams(t *testing.T) {
	tests := []struct {
		name        string
		n           uint64
		p           float64
		wantSize    uint64
		wantHashNum int
	}{
		{"10 elements, 0.001 FP", 10, 0.001, 144, 10},
		{"100 elements, 0.01 FP", 100, 0.01, 959, 7},
		{"500 elements, 0.05 FP", 500, 0.05, 3118, 5},
		{"1000 elements, 0.01 FP", 1000, 0.01, 9586, 7},
		{"2000 elements, 0.01 FP", 2000, 0.01, 19171, 7},
		{"1e6 elements, 0.01 FP", 1e6, 0.01, 9585059, 7},
		{"1e9 elements, 0.01 FP", 1e9, 0.01, 9585058378, 7},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSize, gotHashNum := optimalParams(tt.n, tt.p)
			if gotSize != tt.wantSize {
				t.Errorf("optimalParams() size = %v, want %v", gotSize, tt.wantSize)
			}
			if gotHashNum != tt.wantHashNum {
				t.Errorf("optimalParams() hashes = %v, want %v", gotHashNum, tt.wantHashNum)
			}
		})
	}
}

func TestNewFilter_Invalid(t *testing.T) {
	if NewFilter(0, 0.01) != nil {
		t.Error("NewFilter(0, ...) should return nil")
	}
	if NewFilter(100, 0) != nil {
		t.Error("NewFilter(..., 0) should return nil")
	}
	if NewFilter(100, 1) != nil {
		t.Error("NewFilter(..., 1) should return nil")
	}
	if NewFilter(100, -0.1) != nil {
		t.Error("NewFilter(..., -0.1) should return nil")
	}
}

func TestFilter_AddAndExists(t *testing.T) {
	bf := NewFilter(1000, 0.01)
	if bf == nil {
		t.Fatal("NewFilter returned nil")
	}

	bf.Add([]byte("hello"))
	bf.AddString("world")

	if !bf.Exists([]byte("hello")) {
		t.Error("Exists(hello) should be true")
	}
	if !bf.ExistsString("world") {
		t.Error("ExistsString(world) should be true")
	}
	if bf.Exists([]byte("missing")) {
		t.Error("Exists(missing) should be false")
	}
	if bf.ApproxCount() != 2 {
		t.Errorf("ApproxCount() = %v, want 2", bf.ApproxCount())
	}
}

func TestFilter_FalsePositiveRate(t *testing.T) {
	bf := NewFilter(1000, 0.01)
	if bf.FalsePositiveRate() != 0 {
		t.Error("FP rate should be 0 for empty filter")
	}
	bf.Add([]byte("a"))
	if bf.FalsePositiveRate() <= 0 {
		t.Error("FP rate should increase after insert")
	}
}

func TestFilter_hashWithSeed(t *testing.T) {
	f := &filter{}
	for i := 0; i < 5; i++ {
		got := f.hashWithSeed([]byte("hello"), uint32(i))
		if got == 0 {
			t.Errorf("hashWithSeed(%d) returned 0", i)
		}
	}
}
