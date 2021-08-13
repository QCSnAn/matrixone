package frontend

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"math/rand"
	"matrixone/pkg/defines"
	"strconv"
	"time"
)

// DefaultCapability means default capabilities of the server
var DefaultCapability = CLIENT_LONG_PASSWORD |
	CLIENT_FOUND_ROWS |
	CLIENT_LONG_FLAG |
	CLIENT_CONNECT_WITH_DB |
	CLIENT_LOCAL_FILES |
	CLIENT_PROTOCOL_41 |
	CLIENT_INTERACTIVE |
	CLIENT_TRANSACTIONS |
	CLIENT_SECURE_CONNECTION |
	CLIENT_MULTI_STATEMENTS |
	CLIENT_MULTI_RESULTS |
	CLIENT_PLUGIN_AUTH |
	CLIENT_PLUGIN_AUTH_LENENC_CLIENT_DATA

// DefaultClientConnStatus default server status
var DefaultClientConnStatus = SERVER_STATUS_AUTOCOMMIT

const (
	serverVersion               = "MatrixOne"
	clientProtocolVersion uint8 = 10

	/**
	An answer talks about the charset utf8mb4.
	https://stackoverflow.com/questions/766809/whats-the-difference-between-utf8-general-ci-and-utf8-unicode-ci
	It recommends the charset utf8mb4_0900_ai_ci.
	Maybe we can support utf8mb4_0900_ai_ci in the future.

	A concise research in the Mysql 8.0.23.

	the charset in sever level
	======================================

	mysql> show variables like 'character_set_server';
	+----------------------+---------+
	| Variable_name        | Value   |
	+----------------------+---------+
	| character_set_server | utf8mb4 |
	+----------------------+---------+

	mysql> show variables like 'collation_server';
	+------------------+--------------------+
	| Variable_name    | Value              |
	+------------------+--------------------+
	| collation_server | utf8mb4_0900_ai_ci |
	+------------------+--------------------+

	the charset in database level
	=====================================
	mysql> show variables like 'character_set_database';
	+------------------------+---------+
	| Variable_name          | Value   |
	+------------------------+---------+
	| character_set_database | utf8mb4 |
	+------------------------+---------+

	mysql> show variables like 'collation_database';
	+--------------------+--------------------+
	| Variable_name      | Value              |
	+--------------------+--------------------+
	| collation_database | utf8mb4_0900_ai_ci |
	+--------------------+--------------------+

	*/
	// DefaultCollationID is utf8mb4_bin(46)
	utf8mb4BinCollationID uint8 = 46

	Utf8mb4CollationID uint8 = 45

	AuthNativePassword string = "mysql_native_password"

	// MaxPayloadSize If the payload is larger than or equal to 2^24−1 bytes the length is set to 2^24−1 (ff ff ff)
	//and additional packets are sent with the rest of the payload until the payload of a packet
	//is less than 2^24−1 bytes.
	MaxPayloadSize uint32 = (1 << 24) - 1

	// DefaultMySQLState is the default state of the mySQL
	DefaultMySQLState string = "HY000"

	//for tests
	dumpUser     string = "dump"
	dumpPassword string = "111"
)

type MysqlProtocol struct {
	ProtocolImpl

	//The sequence-id is incremented with each packet and may wrap around.
	//It starts at 0 and is reset to 0 when a new command begins in the Command Phase.
	sequenceId uint8

	//joint capability shared by the server and the client
	capability uint32

	//collation id
	collationID int

	//collation name
	collationName string

	//character set
	charset string

	//max packet size of the client
	maxClientPacketSize uint32

	//the user of the client
	username string

	//the default database for the client
	database string
}

//handshake response 41
type response41 struct {
	capabilities     uint32
	maxPacketSize    uint32
	collationID      uint8
	username         string
	authResponse     []byte
	database         string
	clientPluginName string
}

//handshake response 320
type response320 struct {
	capabilities  uint32
	maxPacketSize uint32
	username      string
	authResponse []byte
	database     string
}

//read an int with length encoded from the buffer at the position
//return the int ; position + the count of bytes for length encoded (1 or 3 or 4 or 9)
func (mp *MysqlProtocol) readIntLenEnc(data []byte, pos int) (uint64, int, bool) {
	if pos >= len(data) {
		return 0, 0, false
	}
	switch data[pos] {
	case 0xfb:
		//zero, one byte
		return 0, pos + 1, true
	case 0xfc:
		// int in two bytes
		if pos+2 >= len(data) {
			return 0, 0, false
		}
		value := uint64(data[pos+1]) |
			uint64(data[pos+2])<<8
		return value, pos + 3, true
	case 0xfd:
		// int in three bytes
		if pos+3 >= len(data) {
			return 0, 0, false
		}
		value := uint64(data[pos+1]) |
			uint64(data[pos+2])<<8 |
			uint64(data[pos+3])<<16
		return value, pos + 4, true
	case 0xfe:
		// int in eight bytes
		if pos+8 >= len(data) {
			return 0, 0, false
		}
		value := uint64(data[pos+1]) |
			uint64(data[pos+2])<<8 |
			uint64(data[pos+3])<<16 |
			uint64(data[pos+4])<<24 |
			uint64(data[pos+5])<<32 |
			uint64(data[pos+6])<<40 |
			uint64(data[pos+7])<<48 |
			uint64(data[pos+8])<<56
		return value, pos + 9, true
	}
	// 0-250
	return uint64(data[pos]), pos + 1, true
}

//write an int with length encoded into the buffer at the position
//return position + the count of bytes for length encoded (1 or 3 or 4 or 9)
func (mp *MysqlProtocol) writeIntLenEnc(data []byte, pos int, value uint64) int {
	switch {
	case value < 251:
		data[pos] = byte(value)
		return pos + 1
	case value < (1 << 16):
		data[pos] = 0xfc
		data[pos+1] = byte(value)
		data[pos+2] = byte(value >> 8)
		return pos + 3
	case value < (1 << 24):
		data[pos] = 0xfd
		data[pos+1] = byte(value)
		data[pos+2] = byte(value >> 8)
		data[pos+3] = byte(value >> 16)
		return pos + 4
	default:
		data[pos] = 0xfe
		data[pos+1] = byte(value)
		data[pos+2] = byte(value >> 8)
		data[pos+3] = byte(value >> 16)
		data[pos+4] = byte(value >> 24)
		data[pos+5] = byte(value >> 32)
		data[pos+6] = byte(value >> 40)
		data[pos+7] = byte(value >> 48)
		data[pos+8] = byte(value >> 56)
		return pos + 9
	}
}

//append an int with length encoded to the buffer
//return the buffer
func (mp *MysqlProtocol) appendIntLenEnc(data []byte, value uint64) []byte {
	tmp := make([]byte, 9)
	pos := mp.writeIntLenEnc(tmp, 0, value)
	return append(data, tmp[:pos]...)
}

//read the count of bytes from the buffer at the position
//return bytes slice ; position + count ; true - succeeded or false - failed
func (mp *MysqlProtocol) readCountOfBytes(data []byte, pos int, count int) ([]byte, int, bool) {
	if pos+count-1 >= len(data) {
		return nil, 0, false
	}
	return data[pos : pos+count], pos + count, true
}

//write the count of bytes into the buffer at the position
//return position + the number of bytes
func (mp *MysqlProtocol) writeCountOfBytes(data []byte, pos int, value []byte) int {
	pos += copy(data[pos:], value)
	return pos
}

//append the count of bytes to the buffer
//return the buffer
func (mp *MysqlProtocol) appendCountOfBytes(data []byte, value []byte) []byte {
	return append(data, value...)
}

//read a string with fixed length from the buffer at the position
//return string ; position + length ; true - succeeded or false - failed
func (mp *MysqlProtocol) readStringFix(data []byte, pos int, length int) (string, int, bool) {
	var sdata []byte
	var ok bool
	sdata, pos, ok = mp.readCountOfBytes(data, pos, length)
	if !ok {
		return "", 0, false
	}
	return string(sdata), pos, true
}

//write a string with fixed length into the buffer at the position
//return pos + string.length
func (mp *MysqlProtocol) writeStringFix(data []byte, pos int, value string, length int) int {
	pos += copy(data[pos:], value[0:length])
	return pos
}

//append a string with fixed length to the buffer
//return the buffer
func (mp *MysqlProtocol) appendStringFix(data []byte, value string, length int) []byte {
	return append(data, []byte(value[:length])...)
}

//read a string appended with zero from the buffer at the position
//return string ; position + length of the string + 1; true - succeeded or false - failed
func (mp *MysqlProtocol) readStringNUL(data []byte, pos int) (string, int, bool) {
	zeroPos := bytes.IndexByte(data[pos:], 0)
	if zeroPos == -1 {
		return "", 0, false
	}
	return string(data[pos : pos+zeroPos]), pos + zeroPos + 1, true
}

//write a string into the buffer at the position, then appended with 0
//return pos + string.length + 1
func (mp *MysqlProtocol) writeStringNUL(data []byte, pos int, value string) int {
	pos = mp.writeStringFix(data, pos, value, len(value))
	data[pos] = 0
	return pos + 1
}

//read a string with length encoded from the buffer at the position
//return string ; position + the count of bytes for length encoded (1 or 3 or 4 or 9) + length of the string; true - succeeded or false - failed
func (mp *MysqlProtocol) readStringLenEnc(data []byte, pos int) (string, int, bool) {
	var value uint64
	var ok bool
	value, pos, ok = mp.readIntLenEnc(data, pos)
	if !ok {
		return "", 0, false
	}
	sLength := int(value)
	if pos+sLength-1 >= len(data) {
		return "", 0, false
	}
	return string(data[pos : pos+sLength]), pos + sLength, true
}

//write a string with length encoded into the buffer at the position
//return position + the count of bytes for length encoded (1 or 3 or 4 or 9) + length of the string;
func (mp *MysqlProtocol) writeStringLenEnc(data []byte, pos int, value string) int {
	pos = mp.writeIntLenEnc(data, pos, uint64(len(value)))
	return mp.writeStringFix(data, pos, value, len(value))
}

//append a string with length encoded to the buffer
//return the buffer
func (mp *MysqlProtocol) appendStringLenEnc(data []byte, value string) []byte {
	data = mp.appendIntLenEnc(data, uint64(len(value)))
	return mp.appendStringFix(data, value, len(value))
}

//append bytes with length encoded to the buffer
//return the buffer
func (mp *MysqlProtocol) appendCountOfBytesLenEnc(data []byte, value []byte) []byte {
	data = mp.appendIntLenEnc(data, uint64(len(value)))
	return mp.appendCountOfBytes(data, value)
}

//append an int64 value converted to string with length encoded to the buffer
//return the buffer
func (mp *MysqlProtocol) appendStringLenEncOfInt64(data []byte, value int64) []byte {
	var tmp []byte
	tmp = strconv.AppendInt(tmp, value, 10)
	return mp.appendCountOfBytesLenEnc(data, tmp)
}

//append an uint64 value converted to string with length encoded to the buffer
//return the buffer
func (mp *MysqlProtocol) appendStringLenEncOfUint64(data []byte, value uint64) []byte {
	var tmp []byte
	tmp = strconv.AppendUint(tmp, value, 10)
	return mp.appendCountOfBytesLenEnc(data, tmp)
}

//append an float32 value converted to string with length encoded to the buffer
//return the buffer
func (mp *MysqlProtocol) appendStringLenEncOfFloat64(data []byte, value float64, bitSize int) []byte {
	var tmp []byte
	tmp = strconv.AppendFloat(tmp, value, 'f', -1, bitSize)
	return mp.appendCountOfBytesLenEnc(data, tmp)
}

//write the count of zeros into the buffer at the position
//return pos + count
func (mp *MysqlProtocol) writeZeros(data []byte, pos int, count int) int {
	for i := 0; i < count; i++ {
		data[pos+i] = 0
	}
	return pos + count
}

//the server calculates the hash value of the password with the algorithm
//and judges it with the authentication data from the client.
//Algorithm: SHA1( password ) XOR SHA1( slat + SHA1( SHA1( password ) ) )
func (mp *MysqlProtocol) checkPassword(password, salt, auth []byte) bool {
	if len(password) == 0 {
		return false
	}
	//hash1 = SHA1(password)
	sha := sha1.New()
	_, err := sha.Write(password)
	if err != nil {
		fmt.Printf("SHA1(password) failed.")
		return false
	}
	hash1 := sha.Sum(nil)

	//hash2 = SHA1(SHA1(password))
	sha.Reset()
	_, err = sha.Write(hash1)
	if err != nil {
		fmt.Printf("SHA1(SHA1(password)) failed.")
		return false
	}
	hash2 := sha.Sum(nil)

	//hash3 = SHA1(salt + SHA1(SHA1(password)))
	sha.Reset()
	_, err = sha.Write(salt)
	if err != nil {
		fmt.Printf("write salt failed.")
		return false
	}
	_, err = sha.Write(hash2)
	if err != nil {
		fmt.Printf("write SHA1(SHA1(password)) failed.")
		return false
	}
	hash3 := sha.Sum(nil)

	//SHA1(password) XOR SHA1(salt + SHA1(SHA1(password)))
	for i := range hash1 {
		hash1[i] ^= hash3[i]
	}

	fmt.Printf("server calculated %v\n", hash1)
	fmt.Printf("client calculated %v\n", auth)

	return bytes.Equal(hash1, auth)
}

//the server authenticate that the client can connect and use the database
func (mp *MysqlProtocol) authenticateUser(authResponse []byte) error {
	//TODO:check the user and the connection
	ses := mp.routine.GetSession()
	//TODO:get the user's password
	var psw []byte
	if mp.username == ses.Pu.SV.GetDumpuser() { //the user dump for test
		psw = []byte(ses.Pu.SV.GetDumppassword())
	}

	//TO Check password
	if mp.checkPassword(psw, mp.salt, authResponse) {
		fmt.Printf("check password succeeded\n")
	} else {
		return fmt.Errorf("check password failed\n")
	}
	return nil
}

func (mp *MysqlProtocol) setSequenceID(value uint8) {
	mp.sequenceId = value
}

//the server makes a handshake v10 packet
//return handshake packet
func (mp *MysqlProtocol) makeHandshakeV10Payload() []byte {
	var data = make([]byte, 256)
	var pos = 0
	//int<1> protocol version
	pos = mp.io.WriteUint8(data, pos, clientProtocolVersion)

	//string[NUL] server version
	pos = mp.writeStringNUL(data, pos, serverVersion)

	//int<4> connection id
	pos = mp.io.WriteUint32(data, pos, mp.connectionID)

	//string[8] auth-plugin-data-part-1
	pos = mp.writeCountOfBytes(data, pos, mp.salt[0:8])

	//int<1> filler 0
	pos = mp.io.WriteUint8(data, pos, 0)

	//int<2>              capabilities flags (lower 2 bytes)
	pos = mp.io.WriteUint16(data, pos, uint16(DefaultCapability&0xFFFF))

	//int<1>              character set
	pos = mp.io.WriteUint8(data, pos, utf8mb4BinCollationID)

	//int<2>              status flags
	pos = mp.io.WriteUint16(data, pos, DefaultClientConnStatus)

	//int<2>              capabilities flags (upper 2 bytes)
	pos = mp.io.WriteUint16(data, pos, uint16((DefaultCapability>>16)&0xFFFF))

	if (DefaultCapability & CLIENT_PLUGIN_AUTH) != 0 {
		//int<1>              length of auth-plugin-data
		//set 21 always
		pos = mp.io.WriteUint8(data, pos, 21)
	} else {
		//int<1>              [00]
		//set 0 always
		pos = mp.io.WriteUint8(data, pos, 0)
	}

	//string[10]     reserved (all [00])
	pos = mp.writeZeros(data, pos, 10)

	if (DefaultCapability & CLIENT_SECURE_CONNECTION) != 0 {
		//string[$len]   auth-plugin-data-part-2 ($len=MAX(13, length of auth-plugin-data - 8))
		pos = mp.writeCountOfBytes(data, pos, mp.salt[8:])
		pos = mp.io.WriteUint8(data, pos, 0)
	}

	if (DefaultCapability & CLIENT_PLUGIN_AUTH) != 0 {
		//string[NUL]    auth-plugin name
		pos = mp.writeStringNUL(data, pos, AuthNativePassword)
	}

	return data[:pos]
}

//the server analyses handshake response41 info from the client
//return true - analysed successfully / false - failed ; response41 ; error
func (mp *MysqlProtocol) analyseHandshakeResponse41(data []byte) (bool, response41, error) {
	var pos = 0
	var ok bool
	var info response41

	//int<4>             capabilities flags of the client, CLIENT_PROTOCOL_41 always set
	info.capabilities, pos, ok = mp.io.ReadUint32(data, pos)
	if !ok {
		return false, info, fmt.Errorf("get capabilities failed")
	}

	if (info.capabilities & CLIENT_PROTOCOL_41) == 0 {
		return false, info, fmt.Errorf("capabilities does not have protocol 41")
	}

	//int<4>             max-packet size
	//max size of a command packet that the client wants to send to the server
	info.maxPacketSize, pos, ok = mp.io.ReadUint32(data, pos)
	if !ok {
		return false, info, fmt.Errorf("get max packet size failed")
	}

	//int<1>             character set
	//connection's default character set
	info.collationID, pos, ok = mp.io.ReadUint8(data, pos)
	if !ok {
		return false, info, fmt.Errorf("get character set failed")
	}

	//string[23]         reserved (all [0])
	//just skip it
	pos += 23

	//string[NUL]        username
	info.username, pos, ok = mp.readStringNUL(data, pos)
	if !ok {
		return false, info, fmt.Errorf("get username failed")
	}

	/*
		if capabilities & CLIENT_PLUGIN_AUTH_LENENC_CLIENT_DATA {
			lenenc-int         length of auth-response
			string[n]          auth-response
		} else if capabilities & CLIENT_SECURE_CONNECTION {
			int<1>             length of auth-response
			string[n]           auth-response
		} else {
			string[NUL]        auth-response
		}
	*/
	if (info.capabilities & CLIENT_PLUGIN_AUTH_LENENC_CLIENT_DATA) != 0 {
		var l uint64
		l, pos, ok = mp.readIntLenEnc(data, pos)
		if !ok {
			return false, info, fmt.Errorf("get length of auth-response failed")
		}
		info.authResponse, pos, ok = mp.readCountOfBytes(data, pos, int(l))
		if !ok {
			return false, info, fmt.Errorf("get auth-response failed")
		}
	} else if (info.capabilities & CLIENT_SECURE_CONNECTION) != 0 {
		var l uint8
		l, pos, ok = mp.io.ReadUint8(data, pos)
		if !ok {
			return false, info, fmt.Errorf("get length of auth-response failed")
		}
		info.authResponse, pos, ok = mp.readCountOfBytes(data, pos, int(l))
		if !ok {
			return false, info, fmt.Errorf("get auth-response failed")
		}
	} else {
		var auth string
		auth, pos, ok = mp.readStringNUL(data, pos)
		if !ok {
			return false, info, fmt.Errorf("get auth-response failed")
		}
		info.authResponse = []byte(auth)
	}

	if (info.capabilities & CLIENT_CONNECT_WITH_DB) != 0 {
		info.database, pos, ok = mp.readStringNUL(data, pos)
		if !ok {
			return false, info, fmt.Errorf("get database failed")
		}
	}

	if (info.capabilities & CLIENT_PLUGIN_AUTH) != 0 {
		info.clientPluginName, pos, ok = mp.readStringNUL(data, pos)
		if !ok {
			return false, info, fmt.Errorf("get auth plugin name failed")
		}

		//to switch authenticate method
		if info.clientPluginName != AuthNativePassword {
			var err error
			if info.authResponse, err = mp.negotiateAuthenticationMethod(); err != nil {
				return false, info, fmt.Errorf("negotiate authentication method failed. error:%v", err)
			}
			info.clientPluginName = AuthNativePassword
		}
	}

	//drop client connection attributes
	return true, info, nil
}

//the server does something after receiving a handshake response41 from the client
//like check user and password
//and other things
func (mp *MysqlProtocol) handleClientResponse41(resp41 response41) error {
	//to do something else
	fmt.Printf("capabilities 0x%x\n", resp41.capabilities)
	fmt.Printf("maxPacketSize %d\n", resp41.maxPacketSize)
	fmt.Printf("collationID %d\n", resp41.collationID)
	fmt.Printf("username %s\n", resp41.username)
	fmt.Printf("authResponse: \n")
	//update the capabilities with client's capabilities
	mp.capability = DefaultCapability & resp41.capabilities

	//character set
	if nameAndCharset, ok := collationID2CharsetAndName[int(resp41.collationID)]; !ok {
		return fmt.Errorf("get collationName and charset failed")
	} else {
		mp.collationID = int(resp41.collationID)
		mp.collationName = nameAndCharset.collationName
		mp.charset = nameAndCharset.charset
	}

	mp.maxClientPacketSize = resp41.maxPacketSize
	mp.username = resp41.username
	mp.database = resp41.database

	fmt.Printf("collationID %d collatonName %s charset %s \n", mp.collationID, mp.collationName, mp.charset)
	fmt.Printf("database %s \n", resp41.database)
	fmt.Printf("clientPluginName %s \n", resp41.clientPluginName)
	return nil
}

//the server analyses handshake response320 info from the old client
//return true - analysed successfully / false - failed ; response320 ; error
func (mp *MysqlProtocol) analyseHandshakeResponse320(data []byte) (bool, response320, error) {
	var pos = 0
	var ok bool
	var info response320
	var capa uint16

	//int<2>             capabilities flags, CLIENT_PROTOCOL_41 never set
	capa, pos, ok = mp.io.ReadUint16(data, pos)
	if !ok {
		return false, info, fmt.Errorf("get capabilities failed")
	}
	info.capabilities = uint32(capa)

	//int<3>             max-packet size
	//max size of a command packet that the client wants to send to the server
	info.maxPacketSize = uint32(data[pos]) | uint32(data[pos+1])<<8 | uint32(data[pos+2])<<16
	pos += 3

	//string[NUL]        username
	info.username, pos, ok = mp.readStringNUL(data, pos)
	if !ok {
		return false, info, fmt.Errorf("get username failed")
	}

	if (info.capabilities & CLIENT_CONNECT_WITH_DB) != 0 {
		var auth string
		auth, pos, ok = mp.readStringNUL(data, pos)
		if !ok {
			return false, info, fmt.Errorf("get auth-response failed")
		}
		info.authResponse = []byte(auth)

		info.database, pos, ok = mp.readStringNUL(data, pos)
		if !ok {
			return false, info, fmt.Errorf("get database failed")
		}
	} else {
		info.authResponse, pos, ok = mp.readCountOfBytes(data, pos, len(data)-pos)
		if !ok {
			return false, info, fmt.Errorf("get auth-response failed")
		}
	}

	return true, info, nil
}

//the server does something after receiving a handshake response320 from the client
//like check user and password
//and other things
func (mp *MysqlProtocol) handleClientResponse320(resp320 response320) error {
	//to do something else
	fmt.Printf("capabilities 0x%x\n", resp320.capabilities)
	fmt.Printf("maxPacketSize %d\n", resp320.maxPacketSize)
	fmt.Printf("username %s\n", resp320.username)
	fmt.Printf("authResponse: \n")

	//update the capabilities with client's capabilities
	mp.capability = DefaultCapability & resp320.capabilities

	//if the client does not notice its default charset, the server gives a default charset.
	//Run the sql in mysql 8.0.23 to get the charset
	//the sql: select * from information_schema.collations where collation_name = 'utf8mb4_general_ci';
	mp.collationID = int(Utf8mb4CollationID)
	mp.collationName = "utf8mb4_general_ci"
	mp.charset = "utf8mb4"

	mp.maxClientPacketSize = resp320.maxPacketSize
	mp.username = resp320.username
	mp.database = resp320.database

	fmt.Printf("collationID %d collatonName %s charset %s \n", mp.collationID, mp.collationName, mp.charset)
	fmt.Printf("database %s \n", resp320.database)
	return nil
}

//the server makes a AuthSwitchRequest that asks the client to authenticate the data with new method
func (mp *MysqlProtocol) makeAuthSwitchRequestPayload(authMethodName string) []byte {
	var data = make([]byte, 1+len(authMethodName)+1+len(mp.salt)+1)
	pos := mp.io.WriteUint8(data, 0, 0xFE)
	pos = mp.writeStringNUL(data, pos, authMethodName)
	pos = mp.writeCountOfBytes(data, pos, mp.salt)
	pos = mp.io.WriteUint8(data, pos, 0)
	return data
}

//the server can send AuthSwitchRequest to ask client to use designated authentication method,
//if both server and client support CLIENT_PLUGIN_AUTH capability.
//return data authenticated with new method
func (mp *MysqlProtocol) negotiateAuthenticationMethod() ([]byte, error) {
	var err error
	aswPkt := mp.makeAuthSwitchRequestPayload(AuthNativePassword)
	err = mp.routine.io.WriteAndFlush(mp.makePackets(aswPkt))
	if err != nil {
		return nil, err
	}

	read, err := mp.routine.io.Read()

	data := read.(*Packet).Payload
	mp.sequenceId++
	return data, nil
}

//make a OK packet
func (mp *MysqlProtocol) makeOKPayload(affectedRows, lastInsertId uint64, statusFlags, warnings uint16, message string) []byte {
	var data = make([]byte, 128)
	var pos = 0
	pos = mp.io.WriteUint8(data, pos, 0)
	pos = mp.writeIntLenEnc(data, pos, affectedRows)
	pos = mp.writeIntLenEnc(data, pos, lastInsertId)
	if (mp.capability & CLIENT_PROTOCOL_41) != 0 {
		pos = mp.io.WriteUint16(data, pos, statusFlags)
		pos = mp.io.WriteUint16(data, pos, warnings)
	} else if (mp.capability & CLIENT_TRANSACTIONS) != 0 {
		pos = mp.io.WriteUint16(data, pos, statusFlags)
	}

	if mp.capability&CLIENT_SESSION_TRACK != 0 {
		//TODO:implement it
	} else {
		alen := Min(len(data)-pos, len(message))
		blen := len(message) - alen
		if alen > 0 {
			pos = mp.writeStringFix(data, pos, message, alen)
		}
		if blen > 0 {
			return mp.appendStringFix(data, message, blen)
		}
		return data[:pos]
	}
	return data
}

//send OK packet to the client
func (mp *MysqlProtocol) makeOKPacket(affectedRows, lastInsertId uint64, status, warnings uint16, message string) []byte {
	okPkt := mp.makeOKPayload(affectedRows, lastInsertId, status, warnings, "")
	return mp.makePackets(okPkt)
}

//make Err packet
func (mp *MysqlProtocol) makeErrPayload(errorCode uint16, sqlState, errorMessage string) []byte {
	var data = make([]byte, 9+len(errorMessage))
	pos := mp.io.WriteUint8(data, 0, 0xff)
	pos = mp.io.WriteUint16(data, pos, errorCode)
	if mp.capability&CLIENT_PROTOCOL_41 != 0 {
		pos = mp.io.WriteUint8(data, pos, '#')
		if len(sqlState) < 5 {
			stuff := "      "
			sqlState += stuff[:5-len(sqlState)]
		}
		pos = mp.writeStringFix(data, pos, sqlState, 5)
	}
	pos = mp.writeStringFix(data, pos, errorMessage, len(errorMessage))
	return data
}

/*
the server sends the Error packet

information from https://dev.mysql.com/doc/mysql-errors/8.0/en/server-error-reference.html
mysql version 8.0.23
usually it is in the directory /usr/local/include/mysql/mysqld_error.h

Error information includes several elements: an error code, SQLSTATE value, and message string.
	Error code: This value is numeric. It is MySQL-specific and is not portable to other database systems.
	SQLSTATE value: This value is a five-character string (for example, '42S02'). SQLSTATE values are taken from ANSI SQL and ODBC and are more standardized than the numeric error codes.
	Message string: This string provides a textual description of the error.
*/
func (mp *MysqlProtocol) makeErrPacket(errorCode uint16, sqlState, errorMessage string) []byte {
	errPkt := mp.makeErrPayload(errorCode, sqlState, errorMessage)
	return mp.makePackets(errPkt)
}

func (mp *MysqlProtocol) makeEOFPayload(warnings, status uint16) []byte {
	data := make([]byte, 10)
	pos := mp.io.WriteUint8(data, 0, 0xFE)
	if mp.capability&CLIENT_PROTOCOL_41 != 0 {
		pos = mp.io.WriteUint16(data, pos, warnings)
		pos = mp.io.WriteUint16(data, pos, status)
	}
	return data[:pos]
}

func (mp *MysqlProtocol) makeEOFPacket(warnings, status uint16) []byte {
	data := mp.makeEOFPayload(warnings, status)
	return mp.makePackets(data)
}

func (mp *MysqlProtocol) MakeEOFPacketIf(warnings, status uint16) []byte {
	//If the CLIENT_DEPRECATE_EOF client capabilities flag is not set, EOF_Packet
	if mp.capability&CLIENT_DEPRECATE_EOF == 0 {
		return mp.makeEOFPacket(warnings, status)
	}
	return nil
}

//the OK or EOF packet
//thread safe
func (mp *MysqlProtocol) MakeEOFOrOkPacket(warnings, status uint16) []byte {
	//If the CLIENT_DEPRECATE_EOF client capabilities flag is set, OK_Packet; else EOF_Packet.
	if mp.capability&CLIENT_DEPRECATE_EOF != 0 {
		return mp.makeOKPacket(0, 0, status, 0, "")
	} else {
		return mp.makeEOFPacket(warnings, status)
	}
}

//make the column information with the format of column definition41
func (mp *MysqlProtocol) makeColumnDefinition41Payload(column *defines.MysqlColumn, cmd int) []byte {
	space := 8*9 + //lenenc bytes of 8 fields
		21 + //fixed-length fields
		3 + // catalog "def"
		len(column.Schema()) +
		len(column.Table()) +
		len(column.OrgTable()) +
		len(column.Name()) +
		len(column.OrgName()) +
		len(column.DefaultValue()) +
		100 // for safe

	data := make([]byte, space)

	//lenenc_str     catalog(always "def")
	pos := mp.writeStringLenEnc(data, 0, "def")

	//lenenc_str     schema
	pos = mp.writeStringLenEnc(data, pos, column.Schema())

	//lenenc_str     table
	pos = mp.writeStringLenEnc(data, pos, column.Table())

	//lenenc_str     org_table
	pos = mp.writeStringLenEnc(data, pos, column.OrgTable())

	//lenenc_str     name
	pos = mp.writeStringLenEnc(data, pos, column.Name())

	//lenenc_str     org_name
	pos = mp.writeStringLenEnc(data, pos, column.OrgName())

	//lenenc_int     length of fixed-length fields [0c]
	pos = mp.io.WriteUint8(data, pos, 0x0c)

	//int<2>              character set
	pos = mp.io.WriteUint16(data, pos, column.Charset())

	//int<4>              column length
	pos = mp.io.WriteUint32(data, pos, column.Length())

	//int<1>              type
	pos = mp.io.WriteUint8(data, pos, column.ColumnType())

	//int<2>              flags
	pos = mp.io.WriteUint16(data, pos, column.Flag())

	//int<1>              decimals
	pos = mp.io.WriteUint8(data, pos, column.Decimal())

	//int<2>              filler [00] [00]
	pos = mp.io.WriteUint16(data, pos, 0)

	if uint8(cmd) == COM_FIELD_LIST {
		pos = mp.writeIntLenEnc(data, pos, uint64(len(column.DefaultValue())))
		pos = mp.writeCountOfBytes(data, pos, column.DefaultValue())
	}

	return data[:pos]
}

// MakeColumnDefinitionPacket the server send the column definition to the client
func (mp *MysqlProtocol) MakeColumnDefinitionPacket(column defines.Column, cmd int) ([]byte, error) {
	mysqlColumn, ok := column.(*defines.MysqlColumn)
	if !ok {
		return nil, fmt.Errorf("sendColumn need MysqlColumn")
	}

	var data []byte
	if mp.capability&CLIENT_PROTOCOL_41 != 0 {
		data = mp.makeColumnDefinition41Payload(mysqlColumn, cmd)
	} else {
		//TODO: ColumnDefinition320
	}

	return mp.makePackets(data), nil
}

// MakeColumnCountPacket makes the column count packet
func (mp *MysqlProtocol) MakeColumnCountPacket(count uint64) []byte {
	data := make([]byte, 20)
	pos := mp.writeIntLenEnc(data, 0, count)

	return mp.makePackets(data[:pos])
}

func (mp *MysqlProtocol) sendColumns(mrs *defines.MysqlResultSet, cmd int, warnings, status uint16) error {
	//column_count * Protocol::ColumnDefinition packets
	for i := uint64(0); i < mrs.GetColumnCount(); i++ {
		var col defines.Column
		col, err := mrs.GetColumn(i)
		if err != nil {
			return err
		}

		packet, err := mp.MakeColumnDefinitionPacket(col, cmd)
		if err != nil {
			return err
		}

		err = mp.routine.io.WriteAndFlush(packet)
		if err != nil {
			return err
		}
	}

	//If the CLIENT_DEPRECATE_EOF client capabilities flag is not set, EOF_Packet
	if mp.capability&CLIENT_DEPRECATE_EOF == 0 {
		err := mp.routine.io.WriteAndFlush(mp.makeEOFPacket(warnings, status))
		if err != nil {
			return err
		}
	}
	return nil
}

//the server convert every row of the result set into the format that mysql protocol needs
func (mp *MysqlProtocol) makeResultSetTextRow(mrs *defines.MysqlResultSet, r uint64) ([]byte, error) {
	var data []byte
	for i := uint64(0); i < mrs.GetColumnCount(); i++ {
		column, err := mrs.GetColumn(i)
		if err != nil {
			return nil, err
		}
		mysqlColumn, ok := column.(*defines.MysqlColumn)
		if !ok {
			return nil, fmt.Errorf("sendColumn need MysqlColumn")
		}

		if isNil, err1 := mrs.ColumnIsNull(r, i); err1 != nil {
			return nil, err1
		} else if isNil {
			//NULL is sent as 0xfb
			data = mp.io.AppendUint8(data, 0xFB)
			continue
		}

		switch mysqlColumn.ColumnType() {
		case defines.MYSQL_TYPE_DECIMAL:
			return nil, fmt.Errorf("unsupported Decimal")
		case defines.MYSQL_TYPE_TINY, defines.MYSQL_TYPE_SHORT, defines.MYSQL_TYPE_INT24, defines.MYSQL_TYPE_LONG, defines.MYSQL_TYPE_YEAR:
			if value, err2 := mrs.GetInt64(r, i); err2 != nil {
				return nil, err2
			} else {
				if mysqlColumn.ColumnType() == defines.MYSQL_TYPE_YEAR {
					if value == 0 {
						data = mp.appendStringLenEnc(data, "0000")
					} else {
						data = mp.appendStringLenEncOfInt64(data, value)
					}
				} else {
					data = mp.appendStringLenEncOfInt64(data, value)
				}
			}
		case defines.MYSQL_TYPE_FLOAT:
			if value, err2 := mrs.GetFloat64(r, i); err2 != nil {
				return nil, err2
			} else {
				data = mp.appendStringLenEncOfFloat64(data, value, 32)
			}
		case defines.MYSQL_TYPE_DOUBLE:
			if value, err2 := mrs.GetFloat64(r, i); err2 != nil {
				return nil, err2
			} else {
				data = mp.appendStringLenEncOfFloat64(data, value, 64)
			}
		case defines.MYSQL_TYPE_LONGLONG:
			if uint32(mysqlColumn.Flag())&defines.UNSIGNED_FLAG != 0 {
				if value, err2 := mrs.GetUint64(r, i); err2 != nil {
					return nil, err2
				} else {
					data = mp.appendStringLenEncOfUint64(data, value)
				}
			} else {
				if value, err2 := mrs.GetInt64(r, i); err2 != nil {
					return nil, err2
				} else {
					data = mp.appendStringLenEncOfInt64(data, value)
				}
			}
		case defines.MYSQL_TYPE_VARCHAR, defines.MYSQL_TYPE_VAR_STRING, defines.MYSQL_TYPE_STRING:
			if value, err2 := mrs.GetString(r, i); err2 != nil {
				return nil, err2
			} else {
				data = mp.appendStringLenEnc(data, value)
			}
		case defines.MYSQL_TYPE_DATE, defines.MYSQL_TYPE_DATETIME, defines.MYSQL_TYPE_TIMESTAMP, defines.MYSQL_TYPE_TIME:
			return nil, fmt.Errorf("unsupported DATE/DATETIME/TIMESTAMP/MYSQL_TYPE_TIME")
		default:
			return nil, fmt.Errorf("unsupported column type %d ", mysqlColumn.ColumnType())
		}
	}
	return data, nil
}

//the server send group row of the result set as an independent packet
//thread safe
func (mp *MysqlProtocol) SendResultSetTextBatchRow(mrs *defines.MysqlResultSet, cnt uint64) error {
	if cnt == 0 {
		return nil
	}

	mp.GetLock().Lock()
	defer mp.GetLock().Unlock()
	var err error = nil

	for i := uint64(0); i < cnt; i++ {
		if err = mp.sendResultSetTextRow(mrs, i); err != nil {
			return err
		}
	}
	return err
}

//the server send every row of the result set as an independent packet
//thread safe
func (mp *MysqlProtocol) SendResultSetTextRow(mrs *defines.MysqlResultSet, r uint64) error {
	mp.GetLock().Lock()
	defer mp.GetLock().Unlock()

	return mp.sendResultSetTextRow(mrs, r)
}

//the server send every row of the result set as an independent packet
func (mp *MysqlProtocol) sendResultSetTextRow(mrs *defines.MysqlResultSet, r uint64) error {
	var data []byte
	var err error
	if data, err = mp.makeResultSetTextRow(mrs, r); err != nil {
		//ERR_Packet in case of error
		err1 := mp.routine.io.WriteAndFlush(mp.makeErrPacket(ER_UNKNOWN_ERROR, DefaultMySQLState, err.Error()))
		if err1 != nil {
			return err1
		}
		return err
	}

	err = mp.routine.io.WriteAndFlush(mp.makePackets(data))
	if err != nil {
		return fmt.Errorf("send result set text row failed. error: %v", err)
	}
	return nil
}

//the server send the result set of execution the client
//the routine follows the article: https://dev.mysql.com/doc/internals/en/com-query-response.html
func (mp *MysqlProtocol) sendResultSet(set defines.ResultSet, cmd int, warnings, status uint16) error {
	mysqlRS, ok := set.(*defines.MysqlResultSet)
	if !ok {
		return fmt.Errorf("sendResultSet need MysqlResultSet")
	}

	//A packet containing a Protocol::LengthEncodedInteger column_count
	err := mp.routine.io.WriteAndFlush(mp.MakeColumnCountPacket(mysqlRS.GetColumnCount()))
	if err != nil {
		return err
	}

	if err = mp.sendColumns(mysqlRS, cmd, warnings, status); err != nil {
		return err
	}

	//One or more ProtocolText::ResultsetRow packets, each containing column_count values
	for i := uint64(0); i < mysqlRS.GetRowCount(); i++ {
		if err = mp.sendResultSetTextRow(mysqlRS, i); err != nil {
			return err
		}
	}

	//If the CLIENT_DEPRECATE_EOF client capabilities flag is set, OK_Packet; else EOF_Packet.
	if mp.capability&CLIENT_DEPRECATE_EOF != 0 {
		err := mp.routine.io.WriteAndFlush(mp.makeOKPacket(0, 0, status, 0, ""))
		if err != nil {
			return err
		}
	} else {
		err := mp.routine.io.WriteAndFlush(mp.makeEOFPacket(warnings, status))
		if err != nil {
			return err
		}
	}

	return nil
}

//the server sends the payload to the client
func (mp *MysqlProtocol) makePackets(data []byte) []byte {
	var packets []byte
	var i = 0
	var length = len(data)
	var curLen = 0
	for ; i < length; i += curLen {
		curLen = Min(int(MaxPayloadSize), length-i)

		//make mysql client protocol header
		//4 bytes
		//int<3>    the length of payload
		var header [4]byte
		mp.io.WriteUint32(header[:], 0, uint32(curLen))

		//int<1> sequence id
		mp.io.WriteUint8(header[:], 3, mp.sequenceId)

		packets = append(packets, header[:]...)
		//send header
		//if err := mp.io.WritePacket(header[:]); err != nil {
		//	return fmt.Errorf("write header failed. error:%v", err)
		//}

		//send payload
		packets = append(packets, data...)
		//if err := mp.io.WritePacket(data[i : i+curLen]); err != nil {
		//	return fmt.Errorf("write payload failed. error:%v", err)
		//}

		mp.sequenceId++

		if i + curLen == length && curLen == int(MaxPayloadSize) {
			//if the size of the last packet is exactly MaxPayloadSize, a zero-size payload should be sent
			header[0] = 0
			header[1] = 0
			header[2] = 0
			header[3] = mp.sequenceId

			//send header / zero-sized packet
			packets = append(packets, header[:]...)
			//if err := mp.io.WritePacket(header[:]); err != nil {
			//	return fmt.Errorf("write header failed. error:%v", err)
			//}

			mp.sequenceId++
		}
	}
	//mp.io.Flush()
	return packets
}

//ther server reads a part of payload from the connection
//the part may be a whole payload
func (mp *MysqlProtocol) recvPartOfPayload() ([]byte, error) {
	//var length int
	//var header []byte
	//var err error
	//if header, err = mp.io.ReadPacket(4); err != nil {
	//	return nil, fmt.Errorf("read header failed.error:%v", err)
	//} else if header[3] != mp.sequenceId {
	//	return nil, fmt.Errorf("client sequence id %d != server sequence id %d", header[3], mp.sequenceId)
	//}

	mp.sequenceId++
	//length = int(uint32(header[0]) | uint32(header[1])<<8 | uint32(header[2])<<16)

	var payload []byte
	//if payload, err = mp.io.ReadPacket(length); err != nil {
	//	return nil, fmt.Errorf("read payload failed.error:%v", err)
	//}
	return payload, nil
}

//the server read a payload from the connection
func (mp *MysqlProtocol) recvPayload() ([]byte, error) {
	payload, err := mp.recvPartOfPayload()
	if err != nil {
		return nil, err
	}

	//only one part
	if len(payload) < int(MaxPayloadSize) {
		return payload, nil
	}

	//payload has been split into many parts.
	//read them all together
	var part []byte
	for {
		part, err = mp.recvPartOfPayload()
		if err != nil {
			return nil, err
		}

		payload = append(payload, part...)

		//only one part
		if len(part) < int(MaxPayloadSize) {
			break
		}
	}
	return payload, nil
}

func NewMysqlClientProtocol(IO IOPackage, connectionID uint32) *MysqlProtocol {
	rand.Seed(time.Now().UTC().UnixNano())
	salt := make([]byte, 20)
	rand.Read(salt)

	mysql := &MysqlProtocol{
		ProtocolImpl: ProtocolImpl{
			io:           IO,
			salt:         salt,
			connectionID: connectionID,
		},
		sequenceId: 0,
		charset:    "utf8mb4",
		capability: DefaultCapability,
	}
	return mysql
}
