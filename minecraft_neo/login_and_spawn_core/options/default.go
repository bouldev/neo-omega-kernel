package options

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	_ "embed"
	"encoding/base64"
	"log"
	rand2 "math/rand"
	"neo-omega-kernel/minecraft/protocol"
	"neo-omega-kernel/minecraft/protocol/login"
	"os"
	"time"

	"github.com/google/uuid"
)

func NewDefaultOptions(
	address, chainData string,
	PrivateKey *ecdsa.PrivateKey,

) *Options {
	opt := &Options{
		ChainData:  chainData,
		Address:    address,
		PrivateKey: PrivateKey,
	}
	opt.Salt = make([]byte, 16)
	_, _ = rand.Read(opt.Salt)
	opt.ErrorLog = log.New(os.Stderr, "", log.LstdFlags)
	defaultIdentityData(&opt.IdentityData)
	defaultClientData(address, opt.IdentityData.DisplayName, &opt.ClientData)
	setAndroidData(&opt.ClientData)
	Request := login.Encode(chainData, opt.ClientData, PrivateKey)
	IdentityData, _, _, _ := login.Parse(Request)
	opt.IdentityData = IdentityData
	opt.Request = Request
	return opt
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
