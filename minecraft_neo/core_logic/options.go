package core_logic

import (
	"bytes"
	"crypto/ecdsa"
	_ "embed"
	"encoding/base64"
	"log"
	"neo-omega-kernel/minecraft/protocol"
	"neo-omega-kernel/minecraft/protocol/login"
	"os"
	"time"

	rand2 "math/rand"

	"github.com/google/uuid"
)

type Options struct {
	ErrorLog             *log.Logger
	DownloadResourcePack func(id uuid.UUID, version string, current, total int) bool

	DisconnectOnUnknownPackets bool
	DisconnectOnInvalidPackets bool
	EnableClientCache          bool
	KeepXBLIdentityData        bool
	Salt                       []byte
	ConnectionRequest          []byte

	IdentityData login.IdentityData
	ClientData   login.ClientData

	PrivateKey *ecdsa.PrivateKey
	Address    string
	ChainData  string
}

// setAndroidData ensures the login.ClientData passed matches settings you would see on an Android device.
func setAndroidData(data *login.ClientData) {
	data.DeviceOS = protocol.DeviceAndroid
	data.GameVersion = protocol.CurrentVersion
}

// defaultIdentityData edits the IdentityData passed to have defaults set to all fields that were left
// unchanged.
func defaultIdentityData(data *login.IdentityData) {
	if data.Identity == "" {
		data.Identity = uuid.New().String()
	}
	if data.DisplayName == "" {
		data.DisplayName = "Steve"
	}
}

//go:embed skin_resource_patch.json
var skinResourcePatch []byte

//go:embed skin_geometry.json
var skinGeometry []byte

// defaultClientData edits the ClientData passed to have defaults set to all fields that were left unchanged.
func defaultClientData(address, username string, d *login.ClientData) {
	rand2.Seed(time.Now().Unix())

	d.ServerAddress = address
	d.ThirdPartyName = username
	if d.DeviceOS == 0 {
		d.DeviceOS = protocol.DeviceAndroid
	}
	if d.GameVersion == "" {
		d.GameVersion = protocol.CurrentVersion
	}
	if d.ClientRandomID == 0 {
		d.ClientRandomID = rand2.Int63()
	}
	if d.DeviceID == "" {
		d.DeviceID = uuid.New().String()
	}
	if d.LanguageCode == "" {
		d.LanguageCode = "en_GB"
	}
	if d.AnimatedImageData == nil {
		d.AnimatedImageData = make([]login.SkinAnimation, 0)
	}
	if d.PersonaPieces == nil {
		d.PersonaPieces = make([]login.PersonaPiece, 0)
	}
	if d.PieceTintColours == nil {
		d.PieceTintColours = make([]login.PersonaPieceTintColour, 0)
	}
	if d.SelfSignedID == "" {
		d.SelfSignedID = uuid.New().String()
	}
	if d.SkinID == "" {
		d.SkinID = uuid.New().String()
	}
	if d.SkinData == "" {
		d.SkinData = base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0, 0, 0, 255}, 32*64))
		d.SkinImageHeight = 32
		d.SkinImageWidth = 64
	}
	if d.SkinResourcePatch == "" {
		d.SkinResourcePatch = base64.StdEncoding.EncodeToString(skinResourcePatch)
	}
	if d.SkinGeometry == "" {
		d.SkinGeometry = base64.StdEncoding.EncodeToString(skinGeometry)
	}
}

func NewDefaultOptions(address string, chainData string, privateKey *ecdsa.PrivateKey) *Options {
	opt := &Options{
		Salt:       make([]byte, 16),
		ErrorLog:   log.New(os.Stderr, "", log.LstdFlags),
		PrivateKey: privateKey,
		ChainData:  chainData,
		Address:    address,
	}
	defaultIdentityData(&opt.IdentityData)
	defaultClientData(address, opt.IdentityData.DisplayName, &opt.ClientData)
	// We login as an Android device and this will show up in the 'titleId' field in the JWT chain, which
	// we can't edit. We just enforce Android data for logging in.
	setAndroidData(&opt.ClientData)
	Request := login.Encode(chainData, opt.ClientData, privateKey)
	opt.ConnectionRequest = Request
	opt.IdentityData, _, _, _ = login.Parse(Request)
	return opt
}
