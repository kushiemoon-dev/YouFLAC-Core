package core

import (
	"errors"
	"testing"
)

func TestClassifyAvailabilityError(t *testing.T) {
	cases := []struct {
		in     error
		reason string
	}{
		{errors.New("Video unavailable. This video has been removed by the uploader"), "removed"},
		{errors.New("Sign in to confirm your age"), "age_restricted"},
		{errors.New("The uploader has not made this video available in your country"), "geo_blocked"},
		{errors.New("This video is private"), "private"},
		{errors.New("random error"), "unknown"},
	}
	for _, c := range cases {
		got := classifyAvailabilityError(c.in)
		if got != c.reason {
			t.Errorf("classifyAvailabilityError(%q) = %q, want %q", c.in.Error(), got, c.reason)
		}
	}
}

func TestCheckAvailable_InvalidURL(t *testing.T) {
	res, err := CheckAvailable("not-a-url")
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
	if res.Available {
		t.Error("expected Available=false")
	}
}
