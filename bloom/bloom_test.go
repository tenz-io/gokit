package bloom

import "testing"

func Test_optimalSize(t *testing.T) {
	type args struct {
		n uint64
		p float64
	}
	tests := []struct {
		name        string
		args        args
		wantSize    uint64
		wantHashNum int
	}{
		{
			name:        "10 elements, 0.01 false positive probability",
			args:        args{n: 10, p: 0.001},
			wantSize:    144,
			wantHashNum: 10,
		},
		{
			name:        "100 elements, 0.01 false positive probability",
			args:        args{n: 100, p: 0.01},
			wantSize:    959,
			wantHashNum: 7,
		},
		{
			name:        "500 elements, 0.05 false positive probability",
			args:        args{n: 500, p: 0.05},
			wantSize:    3118,
			wantHashNum: 5,
		},
		{
			name:        "1000 elements, 0.01 false positive probability",
			args:        args{n: 1000, p: 0.01},
			wantSize:    9586,
			wantHashNum: 7,
		},
		{
			name:        "2000 elements, 0.01 false positive probability",
			args:        args{n: 2000, p: 0.01},
			wantSize:    19171, // ~ 2.34106445 KB
			wantHashNum: 7,
		},
		{
			name:        "millions elements, 0.01 false positive probability",
			args:        args{n: 1e6, p: 0.01},
			wantSize:    9585059, // ~ 1.14903259 MB
			wantHashNum: 7,
		},
		{
			name:        "billions elements, 0.01 false positive probability",
			args:        args{n: 1e9, p: 0.01},
			wantSize:    9585058378, // ~ 1.11584766 GB
			wantHashNum: 7,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSize, gotHashNum := optimalSize(tt.args.n, tt.args.p)
			t.Logf("gotSize: %v, gotHashNum: %v", gotSize, gotHashNum)
			if gotSize != tt.wantSize {
				t.Errorf("optimalSize() gotSize = %v, want %v", gotSize, tt.wantSize)
			}
			if gotHashNum != tt.wantHashNum {
				t.Errorf("optimalSize() gotHashNum = %v, want %v", gotHashNum, tt.wantHashNum)
			}
		})
	}
}

func TestFilter_AddAndExists(t *testing.T) {
	bf := NewFilter(1000, 0.01)

	tests := []struct {
		name     string
		element  []byte
		expected bool
	}{
		{
			name:     "Element 1",
			element:  []byte("element1"),
			expected: true,
		},
		{
			name:     "Element 2",
			element:  []byte("element2"),
			expected: true,
		},
		{
			name:     "Non-existent Element",
			element:  []byte("nonexistent"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expected {
				bf.Add(tt.element)
			}
			exists := bf.Exists(tt.element)
			if exists != tt.expected {
				t.Errorf("Exists() = %v, want %v", exists, tt.expected)
			}
		})
	}
}

func Test_filter_hashWithSeed(t *testing.T) {
	f := &filter{}
	for i := 0; i < 10; i++ {
		got := f.hashWithSeed([]byte("hello"), uint32(i))
		t.Logf("hashWithSeed( %d ) = %v", i, got)
	}
}
