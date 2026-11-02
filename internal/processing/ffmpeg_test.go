package processing

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestTargetHeightDoesNotUpscale(t *testing.T) {
	if targetHeight(720) != 360 || targetHeight(240) != 240 || targetHeight(121) != 120 {
		t.Fatalf("720=%d 240=%d 121=%d", targetHeight(720), targetHeight(240), targetHeight(121))
	}
}

func TestFFmpegRendersOneRenditionWhenPresent(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg ausente")
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("ffprobe ausente")
	}
	dir := t.TempDir()
	source := filepath.Join(dir, "sample.mp4")
	sample := exec.Command("ffmpeg",
		"-hide_banner", "-nostdin", "-y",
		"-f", "lavfi", "-i", "testsrc=size=160x120:rate=1",
		"-f", "lavfi", "-i", "sine=frequency=440:sample_rate=44100",
		"-t", "1", "-shortest",
		"-c:v", "libx264", "-pix_fmt", "yuv420p", "-c:a", "aac",
		source,
	)
	if err := sample.Run(); err != nil {
		t.Fatal(err)
	}
	media := NewFFmpeg("", "")
	probe, err := media.Probe(context.Background(), source)
	if err != nil || !probe.HasVideo || !probe.HasAudio || probe.Height != 120 {
		t.Fatalf("probe=%+v err=%v", probe, err)
	}
	output := filepath.Join(dir, "out")
	if err := os.MkdirAll(output, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := media.Render(context.Background(), source, output, probe); err != nil {
		t.Fatal(err)
	}
	playlist, err := os.ReadFile(filepath.Join(output, "index.m3u8"))
	if err != nil || len(playlist) == 0 {
		t.Fatalf("playlist: %v", err)
	}
	segments, err := filepath.Glob(filepath.Join(output, "segment*.ts"))
	if err != nil || len(segments) == 0 {
		t.Fatalf("segments: %v", err)
	}
	thumb, err := os.ReadFile(filepath.Join(output, "thumbnail.jpg"))
	if err != nil || len(thumb) < 3 || thumb[0] != 0xFF || thumb[1] != 0xD8 {
		t.Fatalf("thumbnail: %v", err)
	}
}
