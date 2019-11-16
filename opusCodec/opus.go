// Package opesCodec is a replacent for the "layeh.com/gumble/opus"
// package used by gumble.
//
// Unlike "layeh.com/gumble/opus", this will not resister itself
// simply by importing it. You must call the Register function
// in order to set configuration values and add the codec to
// grumble.
//
// This package is not thread safe.
package opusCodec

import (
	"gopkg.in/hraban/opus.v2"
	"layeh.com/gumble/gumble"
)

const codecId = 4

type Application = opus.Application

//default parameters for opus
const (
	VoIP     = opus.AppVoIP
	Audio    = opus.AppAudio
	LowDelay = opus.AppRestrictedLowdelay
)

// An AudioCodec that uses opus
type OpusCodec struct {
	application Application
	fec         bool
	packetLoss  int
}

// Sets up codec parameters and registers with grumble.
// You must call this function before
// Sets the basic parameters of the encoder grumble can use opus
func Register(application Application, // Sets default enocding parmeters
	fec bool, // use Forward Error Correction (FEC); good for speach. This is turned on automatically when using opus.AppVoIP
	packetLoss int, // expected packet loss percentage. Higher values reduce total audio quality but allow for much better quality when packets are lost
) {
	codec := &OpusCodec{
		application: application,
		fec:         fec,
		packetLoss:  packetLoss,
	}
	gumble.RegisterAudioCodec(4, codec)
}

// Returns the codec ID as defined by mumble
func (c *OpusCodec) ID() int {
	return codecId
}

// An OpusEncoder designed to work with grumble
type OpusEncoder struct {
	enc         *opus.Encoder
	application opus.Application
	channels    int
}

// An audio decoder for raw opus frames designed to work with grumble
type OpusDecoder struct {
	dec      *opus.Decoder
	channels int
}

// Creates a new OpusEncoderer based on the settings that were used to register the
// package
func (c *OpusCodec) NewEncoder() gumble.AudioEncoder {
	e, _ := opus.NewEncoder(gumble.AudioSampleRate, gumble.AudioChannels, c.application)
	_ = e.SetBitrateToMax()
	_ = e.SetInBandFEC(c.fec)
	_ = e.SetPacketLossPerc(c.packetLoss)
	return &OpusEncoder{
		enc:         e,
		application: c.application,
		channels:    gumble.AudioChannels,
	}
}

// Creates a new OpusDecoder designed to work with grumble
func (c *OpusCodec) NewDecoder() gumble.AudioDecoder {
	d, _ := opus.NewDecoder(gumble.AudioSampleRate, gumble.AudioChannels)
	return &OpusDecoder{
		dec:      d,
		channels: gumble.AudioChannels,
	}
}

// Returns the 'type' of this encoder, as defined by Mumble protocol
func (e *OpusEncoder) ID() int {
	return codecId
}

// Encodes a number of S16LE frames. It is important that the number
// of frames of input is an exact multiple of the frame size * channels
//
// In addition, the frame size must be a supported multiple by a standard
// opus encoder: 2.5, 5, 10, 20, 40, 60ms of audio are the only supported
// types
func (e *OpusEncoder) Encode(pcm []int16, mframeSize, maxDataBytes int) ([]byte, error) {
	buff := make([]byte, maxDataBytes)

	r, err := e.enc.Encode(pcm, buff)
	if err != nil {
		return nil, err
	}
	return buff[0 : r*e.channels], nil
}

// 'Resets' the encoder. In reality it just makes a whole new one
func (e *OpusEncoder) Reset() {
	enc, _ := opus.NewEncoder(gumble.AudioSampleRate, gumble.AudioChannels, e.application)
	_ = enc.SetBitrateToMax()
	fec, _ := enc.InBandFEC()
	enc.SetInBandFEC(fec)
	var packetLoss, _ = enc.PacketLossPerc()
	_ = enc.SetPacketLossPerc(packetLoss)
	e.enc = enc
}

// Returns the 'type' of this codec, as defined by Mumble
func (d *OpusDecoder) ID() int {
	return codecId
}

// Decodes opus data into PCM16_LE raw audio. This currently *does not*
// support FEC features of opus. If any packets are lost, there is little
// that will be done about it
func (d *OpusDecoder) Decode(data []byte, frameSize int) ([]int16, error) {
	var buff = make([]int16, frameSize*d.channels)
	r, err := d.dec.Decode(data, buff)
	if err != nil {
		return nil, err
	}
	return buff[0 : d.channels*r], nil
}

// 'Resets' the encoder. In reality, just deletes it and starts again
func (d *OpusDecoder) Reset() {
	dec, _ := opus.NewDecoder(gumble.AudioSampleRate, gumble.AudioChannels)
	d.dec = dec
}
