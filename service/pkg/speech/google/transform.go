package google

import (
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"github.com/genvmoroz/lale/service/pkg/lang"
	"github.com/genvmoroz/lale/service/pkg/speech"
)

func toSynthesizeSpeechRequest(req speech.ToSpeechRequest) *texttospeechpb.SynthesizeSpeechRequest {
	return &texttospeechpb.SynthesizeSpeechRequest{
		Input:       toSynthesisInput(req.Input),
		Voice:       toVoiceSelectionParams(req.Voice),
		AudioConfig: toAudioConfig(req.AudioConfig),
	}
}

func toSynthesisInput(input string) *texttospeechpb.SynthesisInput {
	return &texttospeechpb.SynthesisInput{
		InputSource: &texttospeechpb.SynthesisInput_Text{
			Text: input,
		},
	}
}

func toVoiceSelectionParams(params speech.VoiceSelectionParams) *texttospeechpb.VoiceSelectionParams {
	return &texttospeechpb.VoiceSelectionParams{
		LanguageCode: params.Language.String(),
		Name:         params.Name,
		SsmlGender:   toGender(params.PreferredVoiceGender),
	}
}

func toAudioConfig(cfg speech.AudioConfig) *texttospeechpb.AudioConfig {
	return &texttospeechpb.AudioConfig{
		AudioEncoding:    toAudioEncoding(cfg.AudioEncoding),
		SpeakingRate:     cfg.SpeakingRate,
		Pitch:            cfg.Pitch,
		VolumeGainDb:     cfg.VolumeGainDb,
		SampleRateHertz:  cfg.SampleRateHertz,
		EffectsProfileId: cfg.EffectsProfileIDs,
	}
}

func toGender(gender speech.VoiceGender) texttospeechpb.SsmlVoiceGender {
	switch gender {
	case speech.Male:
		return texttospeechpb.SsmlVoiceGender_MALE
	case speech.Female:
		return texttospeechpb.SsmlVoiceGender_FEMALE
	case speech.Neutral:
		return texttospeechpb.SsmlVoiceGender_NEUTRAL
	default:
		return texttospeechpb.SsmlVoiceGender_SSML_VOICE_GENDER_UNSPECIFIED
	}
}

func toDomainGender(gender texttospeechpb.SsmlVoiceGender) speech.VoiceGender {
	switch gender {
	case texttospeechpb.SsmlVoiceGender_MALE:
		return speech.Male
	case texttospeechpb.SsmlVoiceGender_FEMALE:
		return speech.Female
	case texttospeechpb.SsmlVoiceGender_NEUTRAL:
		return speech.Neutral
	default:
		return speech.Any
	}
}

func toAudioEncoding(audio speech.AudioEncoding) texttospeechpb.AudioEncoding {
	switch audio {
	case speech.Linear16:
		return texttospeechpb.AudioEncoding_LINEAR16
	case speech.Mp3:
		return texttospeechpb.AudioEncoding_MP3
	case speech.OggOpus:
		return texttospeechpb.AudioEncoding_OGG_OPUS
	case speech.Mulaw:
		return texttospeechpb.AudioEncoding_MULAW
	case speech.Alaw:
		return texttospeechpb.AudioEncoding_ALAW
	default:
		return texttospeechpb.AudioEncoding_AUDIO_ENCODING_UNSPECIFIED
	}
}

func toListVoicesResponse(resp *texttospeechpb.ListVoicesResponse) *speech.ListVoicesResponse {
	if resp == nil {
		return nil
	}

	dResp := speech.ListVoicesResponse{}
	for _, v := range resp.GetVoices() {
		if v == nil {
			continue
		}
		dResp.Voices = append(dResp.Voices, *toVoice(v))
	}
	return &dResp
}

func toVoice(v *texttospeechpb.Voice) *speech.Voice {
	if v == nil {
		return nil
	}

	return &speech.Voice{
		Languages:              toLanguages(v.GetLanguageCodes()),
		Name:                   v.GetName(),
		Gender:                 toDomainGender(v.GetSsmlGender()),
		NaturalSampleRateHertz: v.GetNaturalSampleRateHertz(),
	}
}

func toLanguages(languages []string) []lang.Language {
	if languages == nil {
		return nil
	}

	res := make([]lang.Language, len(languages))
	for i, l := range languages {
		res[i] = lang.Language(l)
	}
	return res
}
