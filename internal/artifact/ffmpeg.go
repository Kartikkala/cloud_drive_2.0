package artifact

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"os/exec"
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

	// TODO 1: GPU detection
	// TODO 2: Generate upto n quality
	// Levels logic from the max quality
	cmd := exec.CommandContext(
		ctx,
		"ffmpeg",
		"-y",
		"-i", filePath,

		// Map video/audio streams for 3 variants
		"-filter_complex",
		"[0:v]split=3[v1][v2][v3];"+
			"[v1]scale=w=640:h=360[v1out];"+
			"[v2]scale=w=842:h=480[v2out];"+
			"[v3]scale=w=1280:h=720[v3out]",

		// 360p
		"-map", "[v1out]",
		"-map", "0:a",
		"-c:v:0", "libx264",
		"-b:v:0", "800k",
		"-c:a:0", "aac",
		"-b:a:0", "96k",

		// 480p
		"-map", "[v2out]",
		"-map", "0:a",
		"-c:v:1", "libx264",
		"-b:v:1", "1400k",
		"-c:a:1", "aac",
		"-b:a:1", "128k",

		// 720p
		"-map", "[v3out]",
		"-map", "0:a",
		"-c:v:2", "libx264",
		"-b:v:2", "2800k",
		"-c:a:2", "aac",
		"-b:a:2", "128k",

		// HLS settings
		"-f", "hls",
		"-hls_time", "6",
		"-hls_playlist_type", "vod",

		// Master playlist
		"-master_pl_name", "master.m3u8",

		// Variant playlists pattern
		"-hls_segment_filename", filepath.Join(outputDir, "%v/segment_%03d.ts"),

		// Stream mapping
		"-var_stream_map", "v:0,a:0 v:1,a:1 v:2,a:2",

		filepath.Join(outputDir, "%v/index.m3u8"),
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
