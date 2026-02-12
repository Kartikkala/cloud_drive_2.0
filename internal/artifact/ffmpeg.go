package artifact

import (
	"context"
	"encoding/json"
	"os/exec"
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

func (svc *Service) ffprobe(
	ctx context.Context,
	FilePath string,
) (*VideoMetadata, error) {
	cmd := exec.CommandContext(ctx,
		"ffprobe",
		"-v", "error",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		FilePath)

	out, err := cmd.Output()

	if err != nil {
		return nil, err
	}

	var ffprobeOutput FFprobeOutput
	var videoMetadata VideoMetadata
	videoMetadata.HasAudio = false
	err = json.Unmarshal(out, &ffprobeOutput)

	if err != nil {
		return nil, err
	}

	for _, stream := range ffprobeOutput.Streams {
		if stream.CodecType == "audio" {
			videoMetadata.HasAudio = true
		} else if stream.CodecType == "video" {
			videoMetadata.Bitrate = ffprobeOutput.Format.Bitrate
			videoMetadata.Codec = stream.CodecName
			videoMetadata.Duration = ffprobeOutput.Format.Duration
			videoMetadata.Height = stream.Height
			videoMetadata.Width = stream.Width
		}
	}

	return &videoMetadata, nil
}

func (svc *Service) ffmpeg(
	ctx context.Context,
	filePath,
	outputDir string,
	duration float64,
	progress func(percent float64),
) error {

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	cmd := exec.CommandContext(
		ctx,
		"ffmpeg",
		"-y",
		"-i", filePath,
		"-c:v", "libx264",
		"-preset", "veryfast",
		"-crf", "23",
		"-c:a", "aac",
		"-b:a", "128k",
		"-f", "hls",
		"-hls_time", "6",
		"-hls_playlist_type", "vod",
		"-hls_segment_filename", filepath.Join(outputDir, "segment_%03d.ts"),
		filepath.Join(outputDir, "index.m3u8"),
	)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(stderr)

	re := regexp.MustCompile(`time=(\d+):(\d+):(\d+\.\d+)`)

	for scanner.Scan() {
		line := scanner.Text()

		match := re.FindStringSubmatch(line)
		if len(match) == 4 {
			h, _ := strconv.ParseFloat(match[1], 64)
			m, _ := strconv.ParseFloat(match[2], 64)
			s, _ := strconv.ParseFloat(match[3], 64)

			current := h*3600 + m*60 + s
			percent := (current / duration) * 100

			progress(percent)
		}
	}

	return cmd.Wait()
}

