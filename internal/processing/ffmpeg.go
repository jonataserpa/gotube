package processing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

// FFmpeg renders one VOD rendition and a JPEG thumbnail with separate arguments.
type FFmpeg struct {
	ffmpeg  string
	ffprobe string
}

// NewFFmpeg uses the binaries on PATH when a path is empty.
func NewFFmpeg(ffmpeg string, ffprobe string) *FFmpeg {
	if ffmpeg == "" {
		ffmpeg = "ffmpeg"
	}
	if ffprobe == "" {
		ffprobe = "ffprobe"
	}
	return &FFmpeg{ffmpeg: ffmpeg, ffprobe: ffprobe}
}

type probeDocument struct {
	Streams []struct {
		CodecType string `json:"codec_type"`
		Height    int    `json:"height"`
	} `json:"streams"`
	Format struct {
		Duration string `json:"duration"`
	} `json:"format"`
}

// Probe reads duration, height and whether audio and video tracks exist.
func (f *FFmpeg) Probe(ctx context.Context, sourcePath string) (Probe, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, f.ffprobe,
		"-v", "error",
		"-print_format", "json",
		"-show_streams",
		"-show_format",
		sourcePath,
	)
	var stdout, stderr cappedBuffer
	stdout.max = 1 << 20
	stderr.max = 4096
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return Probe{}, fmt.Errorf("ffprobe: %w", err)
	}
	var document probeDocument
	if err := json.Unmarshal(stdout.Bytes(), &document); err != nil {
		return Probe{}, err
	}
	var probe Probe
	for _, stream := range document.Streams {
		switch stream.CodecType {
		case "video":
			probe.HasVideo = true
			if stream.Height > probe.Height {
				probe.Height = stream.Height
			}
		case "audio":
			probe.HasAudio = true
		}
	}
	if document.Format.Duration != "" {
		seconds, err := strconv.ParseFloat(document.Format.Duration, 64)
		if err == nil {
			probe.DurationSeconds = int(math.Round(seconds))
		}
	}
	return probe, nil
}

// Render writes index.m3u8, segment*.ts and thumbnail.jpg. Height never increases.
func (f *FFmpeg) Render(ctx context.Context, sourcePath string, outputDir string, probe Probe) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	height := targetHeight(probe.Height)
	args := []string{
		"-hide_banner", "-nostdin", "-y",
		"-i", sourcePath,
		"-vf", fmt.Sprintf("scale=-2:%d", height),
		"-c:v", "libx264", "-preset", "veryfast", "-crf", "28",
		"-pix_fmt", "yuv420p",
	}
	if probe.HasAudio {
		args = append(args, "-c:a", "aac", "-ac", "2")
	} else {
		args = append(args, "-an")
	}
	args = append(args,
		"-hls_time", "4",
		"-hls_playlist_type", "vod",
		"-hls_segment_filename", filepath.Join(outputDir, "segment%03d.ts"),
		filepath.Join(outputDir, "index.m3u8"),
	)
	if err := runBinary(ctx, f.ffmpeg, args); err != nil {
		return err
	}
	return runBinary(ctx, f.ffmpeg, []string{
		"-hide_banner", "-nostdin", "-y",
		"-i", sourcePath,
		"-frames:v", "1",
		"-vf", "scale=640:-2",
		filepath.Join(outputDir, "thumbnail.jpg"),
	})
}

func targetHeight(source int) int {
	height := source
	if height > 360 {
		height = 360
	}
	if height%2 != 0 {
		height--
	}
	if height < 2 {
		return 2
	}
	return height
}

func runBinary(ctx context.Context, bin string, args []string) error {
	cmd := exec.CommandContext(ctx, bin, args...)
	var stderr cappedBuffer
	stderr.max = 4096
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", bin, err)
	}
	return nil
}

type cappedBuffer struct {
	buf bytes.Buffer
	max int
}

func (c *cappedBuffer) Write(p []byte) (int, error) {
	remain := c.max - c.buf.Len()
	if remain > 0 {
		if len(p) > remain {
			_, _ = c.buf.Write(p[:remain])
		} else {
			_, _ = c.buf.Write(p)
		}
	}
	return len(p), nil
}

func (c *cappedBuffer) Bytes() []byte {
	return c.buf.Bytes()
}
