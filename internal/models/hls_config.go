package models

type HLSConfig struct {
	VideoURL      string `json:"video_url"`
	ChunkDuration int    `json:"chunk_duration"`
	AudioChannels int    `json:"audio_channels,omitempty"`
	Resolutions   []int  `json:"resolutions,omitempty"`
}

func (c *HLSConfig) SetDefaults() {
	if c.ChunkDuration == 0 {
		c.ChunkDuration = 10
	}
}
