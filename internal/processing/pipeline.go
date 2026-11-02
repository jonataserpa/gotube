package processing

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5"
)

func (w *Worker) pipeline(ctx context.Context, message uploadedMessage, job claimedJob) (bool, error) {
	key, err := w.sourceKey(ctx, message)
	if err != nil {
		return false, err
	}
	dir, err := os.MkdirTemp("", "gotube-media-")
	if err != nil {
		return false, err
	}
	defer func() { _ = os.RemoveAll(dir) }()
	source := filepath.Join(dir, "source")
	if err := w.objects.Download(ctx, key, source); err != nil {
		return false, err
	}
	probe, err := w.media.Probe(ctx, source)
	if err != nil {
		return false, err
	}
	if !probe.HasVideo || probe.Height < 2 {
		return false, errNoVideo
	}
	output := filepath.Join(dir, "out")
	if err := os.MkdirAll(output, 0o700); err != nil {
		return false, err
	}
	if err := w.media.Render(ctx, source, output, probe); err != nil {
		return false, err
	}
	assets, err := collectAssets(output, message.VideoID, message.Generation)
	if err != nil {
		return false, err
	}
	for _, asset := range assets {
		if err := w.objects.Upload(ctx, asset.key, asset.path, asset.contentType); err != nil {
			return false, err
		}
	}
	if probe.DurationSeconds < 0 {
		probe.DurationSeconds = 0
	}
	return w.finish(ctx, job, probe.DurationSeconds, assets)
}

func (w *Worker) sourceKey(ctx context.Context, message uploadedMessage) (string, error) {
	var key string
	err := w.pool.QueryRow(ctx, `
		SELECT object_key
		FROM upload_sessions
		WHERE id = $1 AND video_id = $2 AND completed_at IS NOT NULL
	`, message.UploadSessionID, message.VideoID).Scan(&key)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", errInvalidOutput
	}
	if err != nil {
		return "", err
	}
	return key, nil
}

func collectAssets(dir string, videoID string, generation int) ([]storedAsset, error) {
	prefix := fmt.Sprintf("videos/%s/generations/%d", videoID, generation)
	playlistPath := filepath.Join(dir, "index.m3u8")
	playlist, err := os.ReadFile(playlistPath)
	if err != nil || !strings.Contains(string(playlist), "#EXTM3U") {
		return nil, errInvalidOutput
	}
	names, err := filepath.Glob(filepath.Join(dir, "segment*.ts"))
	if err != nil || len(names) == 0 {
		return nil, errInvalidOutput
	}
	thumbPath := filepath.Join(dir, "thumbnail.jpg")
	thumb, err := os.ReadFile(thumbPath)
	if err != nil || len(thumb) < 3 || thumb[0] != 0xFF || thumb[1] != 0xD8 {
		return nil, errInvalidOutput
	}
	var assets []storedAsset
	for _, name := range names {
		item, err := newAsset("HLS_SEGMENT", prefix+"/hls/"+filepath.Base(name), "video/mp2t", name)
		if err != nil {
			return nil, err
		}
		assets = append(assets, item)
	}
	playlistAsset, err := newAsset("HLS_PLAYLIST", prefix+"/hls/index.m3u8", "application/vnd.apple.mpegurl", playlistPath)
	if err != nil {
		return nil, err
	}
	thumbAsset, err := newAsset("THUMBNAIL", prefix+"/thumbnail.jpg", "image/jpeg", thumbPath)
	if err != nil {
		return nil, err
	}
	return append(assets, playlistAsset, thumbAsset), nil
}

func newAsset(kind string, key string, contentType string, path string) (storedAsset, error) {
	sum, err := fileChecksum(path)
	if err != nil || sum == "" {
		return storedAsset{}, errInvalidOutput
	}
	return storedAsset{kind: kind, key: key, contentType: contentType, path: path, checksum: sum}, nil
}
