package uqholder

import (
	"bytes"
	"errors"
	"neo-omega-kernel/minecraft/protocol"
	"neo-omega-kernel/neomega/encoding/binary_read_write"
	"neo-omega-kernel/neomega/encoding/little_endian"
	"time"
)

func (p *Player) Marshal() (data []byte, err error) {
	basicWriter := bytes.NewBuffer(nil)
	writer := binary_read_write.WrapBinaryWriter(basicWriter)
	err = writer.Write(p.UUID[:])
	if err != nil {
		return nil, err
	}
	err = writer.WriteByte(boolToByte(p.knownUUID))
	if err != nil {
		return nil, err
	}
	err = little_endian.WriteInt64(writer, p.EntityUniqueID)
	if err != nil {
		return nil, err
	}
	err = writer.WriteByte(boolToByte(p.knownEntityUniqueID))
	if err != nil {
		return nil, err
	}
	err = little_endian.WriteInt64(writer, p.LoginTime.Unix())
	if err != nil {
		return nil, err
	}
	err = writer.WriteByte(boolToByte(p.knownLoginTime))
	if err != nil {
		return nil, err
	}
	err = little_endian.WriteString(writer, p.Username)
	if err != nil {
		return nil, err
	}
	err = writer.WriteByte(boolToByte(p.knownUsername))
	if err != nil {
		return nil, err
	}
	err = little_endian.WriteString(writer, p.PlatformChatID)
	if err != nil {
		return nil, err
	}
	err = writer.WriteByte(boolToByte(p.knownPlatformChatID))
	if err != nil {
		return nil, err
	}
	err = little_endian.WriteInt32(writer, p.BuildPlatform)
	if err != nil {
		return nil, err
	}
	err = writer.WriteByte(boolToByte(p.knownBuildPlatform))
	if err != nil {
		return nil, err
	}
	err = little_endian.WriteString(writer, p.SkinID)
	if err != nil {
		return nil, err
	}
	err = writer.WriteByte(boolToByte(p.knownSkinID))
	if err != nil {
		return nil, err
	}
	err = little_endian.WriteUint32(writer, p.PropertiesFlag)
	if err != nil {
		return nil, err
	}
	err = writer.WriteByte(boolToByte(p.knownPropertiesFlag))
	if err != nil {
		return nil, err
	}
	err = little_endian.WriteUint32(writer, p.CommandPermissionLevel)
	if err != nil {
		return nil, err
	}
	err = writer.WriteByte(boolToByte(p.knownCommandPermissionLevel))
	if err != nil {
		return nil, err
	}
	err = little_endian.WriteUint32(writer, p.ActionPermissions)
	if err != nil {
		return nil, err
	}
	err = writer.WriteByte(boolToByte(p.knownActionPermissions))
	if err != nil {
		return nil, err
	}
	err = little_endian.WriteUint32(writer, p.OPPermissionLevel)
	if err != nil {
		return nil, err
	}
	err = writer.WriteByte(boolToByte(p.knownOPPermissionLevel))
	if err != nil {
		return nil, err
	}
	err = little_endian.WriteUint32(writer, p.CustomStoredPermissions)
	if err != nil {
		return nil, err
	}
	err = writer.WriteByte(boolToByte(p.knownCustomStoredPermissions))
	if err != nil {
		return nil, err
	}
	err = little_endian.WriteString(writer, p.DeviceID)
	if err != nil {
		return nil, err
	}
	err = writer.WriteByte(boolToByte(p.knownDeviceID))
	if err != nil {
		return nil, err
	}
	err = little_endian.WriteUint64(writer, p.EntityRuntimeID)
	if err != nil {
		return nil, err
	}
	err = writer.WriteByte(boolToByte(p.knownEntityRuntimeID))
	if err != nil {
		return nil, err
	}
	protocol.NewWriter(basicWriter, 0).EntityMetadata(&p.EntityMetadata)
	err = writer.WriteByte(boolToByte(p.knownEntityMetadata))
	if err != nil {
		return nil, err
	}
	err = writer.WriteByte(boolToByte(p.Online))
	if err != nil {
		return nil, err
	}
	return basicWriter.Bytes(), err
}

func boolToByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}

func byteToBool(b byte) (bool, error) {
	if b == 1 {
		return true, nil
	}
	if b == 0 {
		return false, nil
	}
	return false, errors.New("byteToBool: invalid byte")
}

func (p *Player) Unmarshal(data []byte) (err error) {
	basicReader := bytes.NewReader(data)
	reader := binary_read_write.WrapBinaryReader(basicReader)
	readAndGetBool := func() (bool, error) {
		b, err := reader.ReadByte()
		if err != nil {
			return false, err
		}
		return byteToBool(b)
	}

	err = reader.Read(p.UUID[:])
	if err != nil {
		return err
	}
	p.knownUUID, err = readAndGetBool()
	if err != nil {
		return err
	}
	p.EntityUniqueID, err = little_endian.Int64(reader)
	if err != nil {
		return err
	}
	p.knownEntityUniqueID, err = readAndGetBool()
	if err != nil {
		return err
	}
	loginTime, err := little_endian.Int64(reader)
	if err != nil {
		return err
	}
	p.LoginTime = time.Unix(loginTime, 0)
	p.knownLoginTime, err = readAndGetBool()
	if err != nil {
		return err
	}
	p.Username, err = little_endian.String(reader)
	if err != nil {
		return err
	}
	p.knownUsername, err = readAndGetBool()
	if err != nil {
		return err
	}
	p.PlatformChatID, err = little_endian.String(reader)
	if err != nil {
		return err
	}
	p.knownPlatformChatID, err = readAndGetBool()
	if err != nil {
		return err
	}
	p.BuildPlatform, err = little_endian.Int32(reader)
	if err != nil {
		return err
	}
	p.knownBuildPlatform, err = readAndGetBool()
	if err != nil {
		return err
	}
	p.SkinID, err = little_endian.String(reader)
	if err != nil {
		return err
	}
	p.knownSkinID, err = readAndGetBool()
	if err != nil {
		return err
	}
	p.PropertiesFlag, err = little_endian.Uint32(reader)
	if err != nil {
		return err
	}
	p.knownPropertiesFlag, err = readAndGetBool()
	if err != nil {
		return err
	}
	p.CommandPermissionLevel, err = little_endian.Uint32(reader)
	if err != nil {
		return err
	}
	p.knownCommandPermissionLevel, err = readAndGetBool()
	if err != nil {
		return err
	}
	p.ActionPermissions, err = little_endian.Uint32(reader)
	if err != nil {
		return err
	}
	p.knownActionPermissions, err = readAndGetBool()
	if err != nil {
		return err
	}
	p.OPPermissionLevel, err = little_endian.Uint32(reader)
	if err != nil {
		return err
	}
	p.knownOPPermissionLevel, err = readAndGetBool()
	if err != nil {
		return err
	}
	p.CustomStoredPermissions, err = little_endian.Uint32(reader)
	if err != nil {
		return err
	}
	p.knownCustomStoredPermissions, err = readAndGetBool()
	if err != nil {
		return err
	}
	p.DeviceID, err = little_endian.String(reader)
	if err != nil {
		return err
	}
	p.knownDeviceID, err = readAndGetBool()
	if err != nil {
		return err
	}
	p.EntityRuntimeID, err = little_endian.Uint64(reader)
	if err != nil {
		return err
	}
	p.knownEntityRuntimeID, err = readAndGetBool()
	if err != nil {
		return err
	}
	protocol.NewReader(basicReader, 0).EntityMetadata(&p.EntityMetadata)
	p.knownEntityMetadata, err = readAndGetBool()
	if err != nil {
		return err
	}
	p.Online, err = readAndGetBool()
	if err != nil {
		return err
	}
	return nil
}
