package speech

import "github.com/genvmoroz/lale/service/pkg/lang"

type (
	ToSpeechRequest struct {
		Input       string
		Voice       VoiceSelectionParams
		AudioConfig AudioConfig
	}

	VoiceSelectionParams struct {
		Language             lang.Language
		Name                 string
		PreferredVoiceGender VoiceGender
	}

	AudioConfig struct {
		AudioEncoding     AudioEncoding
		SpeakingRate      float64
		Pitch             float64
		VolumeGainDb      float64
		SampleRateHertz   int32
		EffectsProfileIDs []string
	}

	ListVoicesResponse struct {
		Voices []Voice
	}

	Voice struct {
		Languages              []lang.Language
		Name                   string
		Gender                 VoiceGender
		NaturalSampleRateHertz int32
	}
)

type VoiceGender int32

const (
	Any VoiceGender = iota
	Male
	Female
	Neutral
)

type AudioEncoding int32

const (
	Unknown AudioEncoding = iota
	Linear16
	Mp3
	OggOpus
	Mulaw
	Alaw
)
