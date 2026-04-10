package core

import (
	"context"
	"testing"
	"time"
)

func TestFetchChannelUploads_AllItems(t *testing.T) {
	SetYtdlpBinaryForTests("./testdata/ytdlp_channel_ok.sh")
	t.Cleanup(func() { SetYtdlpBinaryForTests("yt-dlp") })

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	items, errs := FetchChannelUploads(ctx, "https://www.youtube.com/@Test", ChannelOpts{})
	var got []VideoInfoLite
	for v := range items {
		got = append(got, v)
	}
	if err := <-errs; err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 10 {
		t.Errorf("expected 10 items, got %d", len(got))
	}
}

func TestFetchChannelUploads_OnlyLongForm(t *testing.T) {
	SetYtdlpBinaryForTests("./testdata/ytdlp_channel_ok.sh")
	t.Cleanup(func() { SetYtdlpBinaryForTests("yt-dlp") })

	ctx := context.Background()
	items, errs := FetchChannelUploads(ctx, "https://www.youtube.com/@Test", ChannelOpts{OnlyLongForm: true})
	var got []VideoInfoLite
	for v := range items {
		got = append(got, v)
	}
	<-errs
	if len(got) != 8 {
		t.Errorf("expected 8 longform items, got %d", len(got))
	}
	for _, v := range got {
		if v.IsShort {
			t.Errorf("expected no shorts, got short: %+v", v)
		}
	}
}

func TestFetchChannelUploads_MaxItems(t *testing.T) {
	SetYtdlpBinaryForTests("./testdata/ytdlp_channel_ok.sh")
	t.Cleanup(func() { SetYtdlpBinaryForTests("yt-dlp") })

	ctx := context.Background()
	items, errs := FetchChannelUploads(ctx, "https://www.youtube.com/@Test", ChannelOpts{MaxItems: 3})
	var got []VideoInfoLite
	for v := range items {
		got = append(got, v)
	}
	<-errs
	if len(got) != 3 {
		t.Errorf("expected 3 items (MaxItems cap), got %d", len(got))
	}
}

func TestFetchChannelUploads_MalformedLineSkipped(t *testing.T) {
	SetYtdlpBinaryForTests("./testdata/ytdlp_channel_mixed.sh")
	t.Cleanup(func() { SetYtdlpBinaryForTests("yt-dlp") })

	ctx := context.Background()
	items, errs := FetchChannelUploads(ctx, "https://www.youtube.com/@Test", ChannelOpts{})
	var got []VideoInfoLite
	for v := range items {
		got = append(got, v)
	}
	if err := <-errs; err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 {
		t.Errorf("expected 3 valid items (1 malformed skipped), got %d", len(got))
	}
}

func TestFetchChannelUploads_CancelCtx(t *testing.T) {
	SetYtdlpBinaryForTests("./testdata/ytdlp_channel_ok.sh")
	t.Cleanup(func() { SetYtdlpBinaryForTests("yt-dlp") })

	ctx, cancel := context.WithCancel(context.Background())
	items, errs := FetchChannelUploads(ctx, "https://www.youtube.com/@Test", ChannelOpts{})

	// Read one item then cancel
	<-items
	cancel()

	// Drain remaining
	for range items {}
	<-errs // must not block
}
