/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package sampling

import "testing"

func TestRandomFloat64Range(t *testing.T) {
	for i := 0; i < 10000; i++ {
		v := randomFloat64()
		if v < 0.0 || v >= 1.0 {
			t.Fatalf("value out of range [0.0, 1.0): %v", v)
		}
	}
}

func TestShouldSampleWithRate(t *testing.T) {
	s := NewSampler(nil)

	if !s.shouldSampleWithRate(1.0) {
		t.Error("rate 1.0 must always sample")
	}
	if s.shouldSampleWithRate(0.0) {
		t.Error("rate 0.0 must never sample")
	}

	sampled := 0
	for i := 0; i < 10000; i++ {
		if s.shouldSampleWithRate(0.5) {
			sampled++
		}
	}
	if sampled < 4000 || sampled > 6000 {
		t.Errorf("rate 0.5 sampled %d of 10000 events", sampled)
	}
}
