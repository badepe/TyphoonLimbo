package main

import (
	"fmt"
	"github.com/TyphoonMC/go.uuid"
	"log"
)

type PacketHandshake struct {
	protocol Protocol
	address  string
	port     uint16
	state    State
}

func (packet *PacketHandshake) Read(player *Player, length int) (err error) {
	protocol, err := player.ReadVarInt()
	if err != nil {
		log.Print(err)
		return
	}
	packet.protocol = Protocol(protocol)
	packet.address, err = player.ReadStringLimited(config.BufferConfig.HandshakeAddress)
	if err != nil {
		log.Print(err)
		return
	}
	packet.port, err = player.ReadUInt16()
	if err != nil {
		log.Print(err)
		return
	}
	state, err := player.ReadVarInt()
	if err != nil {
		log.Print(err)
		return
	}
	packet.state = State(state)
	return
}
func (packet *PacketHandshake) Write(player *Player) (err error) {
	return
}
func (packet *PacketHandshake) Handle(player *Player) {
	player.state = packet.state
	player.protocol = packet.protocol
	player.inaddr.address = packet.address
	player.inaddr.port = packet.port
}
func (packet *PacketHandshake) Id() int {
	return 0x00
}

type PacketStatusRequest struct{}

func (packet *PacketStatusRequest) Read(player *Player, length int) (err error) {
	return
}
func (packet *PacketStatusRequest) Write(player *Player) (err error) {
	return
}
func (packet *PacketStatusRequest) Handle(player *Player) {
	protocol := COMPATIBLE_PROTO[0]
	if IsCompatible(player.protocol) {
		protocol = player.protocol
	}

	max_players := config.MaxPlayers
	motd := config.Motd

	if max_players < players_count && !config.Restricted {
		max_players = players_count
	}

	response := PacketStatusResponse{
		response: fmt.Sprintf(`{"version":{"name":"Typhoon","protocol":%d},"players":{"max":%d,"online":%d,"sample":[]},"description":{"text":"%s"},"favicon":"%s","modinfo":{"type":"FML","modList":[]}}`, protocol, max_players, players_count, JsonEscape(motd), JsonEscape(favicon)),
	}
	player.WritePacket(&response)
}
func (packet *PacketStatusRequest) Id() int {
	return 0x00
}

type PacketStatusResponse struct {
	response string
}

func (packet *PacketStatusResponse) Read(player *Player, length int) (err error) {
	return
}
func (packet *PacketStatusResponse) Write(player *Player) (err error) {
	err = player.WriteString(packet.response)
	if err != nil {
		log.Print(err)
		return
	}
	return
}
func (packet *PacketStatusResponse) Handle(player *Player) {}
func (packet *PacketStatusResponse) Id() int {
	return 0x00
}

type PacketStatusPing struct {
	time uint64
}

func (packet *PacketStatusPing) Read(player *Player, length int) (err error) {
	packet.time, err = player.ReadUInt64()
	if err != nil {
		log.Print(err)
		return
	}
	return
}
func (packet *PacketStatusPing) Write(player *Player) (err error) {
	err = player.WriteUInt64(packet.time)
	if err != nil {
		log.Print(err)
		return
	}
	return
}
func (packet *PacketStatusPing) Handle(player *Player) {
	player.WritePacket(packet)
}
func (packet *PacketStatusPing) Id() int {
	return 0x01
}

type PacketLoginStart struct {
	username string
}

func (packet *PacketLoginStart) Read(player *Player, length int) (err error) {
	packet.username, err = player.ReadStringLimited(config.BufferConfig.PlayerName)
	if err != nil {
		log.Print(err)
		return
	}
	return
}
func (packet *PacketLoginStart) Write(player *Player) (err error) {
	return
}

var (
	join_game = PacketPlayJoinGame{
		entity_id:     0,
		gamemode:      SPECTATOR,
		dimension:     END,
		difficulty:    NORMAL,
		level_type:    DEFAULT,
		max_players:   0xFF,
		reduced_debug: false,
	}
	position_look = PacketPlayerPositionLook{}
)

func (packet *PacketLoginStart) Handle(player *Player) {
	if !IsCompatible(player.protocol) {
		player.LoginKick("Incompatible version")
		return
	}

	max_players := config.MaxPlayers

	if max_players <= players_count && config.Restricted {
		player.LoginKick("Server is full")
	}

	player.name = packet.username

	if config.Compression && player.protocol >= V1_8 {
		setCompression := PacketSetCompression{config.Threshold}
		player.WritePacket(&setCompression)
		player.compression = true
	}

	success := PacketLoginSuccess{
		uuid:     player.uuid,
		username: player.name,
	}
	player.WritePacket(&success)
	player.state = PLAY
	player.register()

	player.WritePacket(&join_game)
	player.WritePacket(&position_look)

	if &join_message != nil {
		player.WritePacket(&join_message)
	}
	if &bossbar_create != nil {
		player.WritePacket(&bossbar_create)
	}
	if &playerlist_hf != nil {
		player.WritePacket(&playerlist_hf)
	}
}
func (packet *PacketLoginStart) Id() int {
	return 0x00
}

type PacketLoginDisconnect struct {
	component string
}

func (packet *PacketLoginDisconnect) Read(player *Player, length int) (err error) {
	return
}
func (packet *PacketLoginDisconnect) Write(player *Player) (err error) {
	err = player.WriteString(packet.component)
	if err != nil {
		log.Print(err)
		return
	}
	return
}
func (packet *PacketLoginDisconnect) Handle(player *Player) {}
func (packet *PacketLoginDisconnect) Id() int {
	return 0x00
}

type PacketLoginSuccess struct {
	uuid     string
	username string
}

func (packet *PacketLoginSuccess) Read(player *Player, length int) (err error) {
	return
}
func (packet *PacketLoginSuccess) Write(player *Player) (err error) {
	err = player.WriteString(packet.uuid)
	if err != nil {
		log.Print(err)
		return
	}
	err = player.WriteString(packet.username)
	if err != nil {
		log.Print(err)
		return
	}
	return
}
func (packet *PacketLoginSuccess) Handle(player *Player) {}
func (packet *PacketLoginSuccess) Id() int {
	return 0x02
}

type PacketSetCompression struct {
	threshold int
}

func (packet *PacketSetCompression) Read(player *Player, length int) (err error) {
	return
}
func (packet *PacketSetCompression) Write(player *Player) (err error) {
	err = player.WriteVarInt(packet.threshold)
	if err != nil {
		log.Print(err)
		return
	}
	return
}
func (packet *PacketSetCompression) Handle(player *Player) {}
func (packet *PacketSetCompression) Id() int {
	return 0x03
}

type PacketPlayChat struct {
	message string
}

func (packet *PacketPlayChat) Read(player *Player, length int) (err error) {
	packet.message, err = player.ReadStringLimited(config.BufferConfig.ChatMessage)
	if err != nil {
		log.Print(err)
		return
	}
	return
}
func (packet *PacketPlayChat) Write(player *Player) (err error) {
	return
}
func (packet *PacketPlayChat) Handle(player *Player) {
	if len(packet.message) > 0 && packet.message[0] != '/' {
		player.WritePacket(&PacketPlayMessage{
			component: fmt.Sprintf(`{"text":"<%s> %s"}`, player.name, JsonEscape(packet.message)),
			position:  CHAT_BOX,
		})
	}
}
func (packet *PacketPlayChat) Id() int {
	return 0x02
}

type PacketPlayMessage struct {
	component string
	position  ChatPosition
}

func (packet *PacketPlayMessage) Read(player *Player, length int) (err error) {
	return
}
func (packet *PacketPlayMessage) Write(player *Player) (err error) {
	err = player.WriteString(packet.component)
	if err != nil {
		log.Print(err)
		return
	}
	if player.protocol > V1_7_6 {
		err = player.WriteUInt8(uint8(packet.position))
		if err != nil {
			log.Print(err)
			return
		}
	}
	return
}
func (packet *PacketPlayMessage) Handle(player *Player) {}
func (packet *PacketPlayMessage) Id() int {
	return 0x0F
}

type PacketBossBar struct {
	uuid     uuid.UUID
	action   BossBarAction
	title    string
	health   float32
	color    BossBarColor
	division BossBarDivision
	flags    uint8
}

func (packet *PacketBossBar) Read(player *Player, length int) (err error) {
	return
}
func (packet *PacketBossBar) Write(player *Player) (err error) {
	err = player.WriteUUID(packet.uuid)
	if err != nil {
		log.Print(err)
		return
	}
	err = player.WriteVarInt(int(packet.action))
	if err != nil {
		log.Print(err)
		return
	}
	if packet.action == BOSSBAR_UPDATE_TITLE || packet.action == BOSSBAR_ADD {
		err = player.WriteString(packet.title)
		if err != nil {
			log.Print(err)
			return
		}
	}
	if packet.action == BOSSBAR_UPDATE_HEALTH || packet.action == BOSSBAR_ADD {
		err = player.WriteFloat32(packet.health)
		if err != nil {
			log.Print(err)
			return
		}
	}
	if packet.action == BOSSBAR_UPDATE_STYLE || packet.action == BOSSBAR_ADD {
		err = player.WriteVarInt(int(packet.color))
		if err != nil {
			log.Print(err)
			return
		}
		err = player.WriteVarInt(int(packet.division))
		if err != nil {
			log.Print(err)
			return
		}
	}
	if packet.action == BOSSBAR_UPDATE_STYLE || packet.action == BOSSBAR_ADD {
		err = player.WriteUInt8(packet.flags)
		if err != nil {
			log.Print(err)
			return
		}
	}
	return
}
func (packet *PacketBossBar) Handle(player *Player) {}
func (packet *PacketBossBar) Id() int {
	return 0x0C
}

type PacketPlayPluginMessage struct {
	channel string
	data    []byte
}

func (packet *PacketPlayPluginMessage) Read(player *Player, length int) (err error) {
	var read int
	packet.channel, read, err = player.ReadNStringLimited(20)
	if err != nil {
		log.Print(err)
		return
	}

	dataLength := length - read
	if player.protocol < V1_8 {
		sread, err := player.ReadUInt16()
		if err != nil {
			log.Print(err)
			return err
		}
		dataLength = int(sread)
	}

	packet.data, err = player.ReadByteArray(dataLength)
	if err != nil {
		log.Print(err)
		return
	}
	return
}
func (packet *PacketPlayPluginMessage) Write(player *Player) (err error) {
	err = player.WriteString(packet.channel)
	if err != nil {
		log.Print(err)
		return
	}
	if player.protocol < V1_8 {
		err = player.WriteUInt16(uint16(len(packet.data)))
		if err != nil {
			log.Print(err)
			return err
		}
	}
	err = player.WriteByteArray(packet.data)
	if err != nil {
		log.Print(err)
		return
	}
	return
}
func (packet *PacketPlayPluginMessage) Handle(player *Player) {
	if packet.channel == "MC|Brand" || packet.channel == "minecraft:brand" {
		log.Printf("%s is using %s client", player.name, string(packet.data))
		player.WritePacket(&PacketPlayPluginMessage{
			packet.channel,
			[]byte("typhoonlimbo"),
		})
	}
}
func (packet *PacketPlayPluginMessage) Id() int {
	return 0x18
}

type PacketPlayDisconnect struct {
	component string
}

func (packet *PacketPlayDisconnect) Read(player *Player, length int) (err error) {
	return
}
func (packet *PacketPlayDisconnect) Write(player *Player) (err error) {
	err = player.WriteString(packet.component)
	if err != nil {
		log.Print(err)
		return
	}
	return
}
func (packet *PacketPlayDisconnect) Handle(player *Player) {}
func (packet *PacketPlayDisconnect) Id() int {
	return 0x1A
}

type PacketPlayKeepAlive struct {
	id int
}

func (packet *PacketPlayKeepAlive) Read(player *Player, length int) (err error) {
	if player.protocol >= V1_12_2 {
		id, stt := player.ReadUInt64()
		packet.id = int(id)
		err = stt
	} else if player.protocol <= V1_7_6 {
		id, stt := player.ReadUInt32()
		packet.id = int(id)
		err = stt
	} else {
		packet.id, err = player.ReadVarInt()
	}
	if err != nil {
		log.Print(err)
		return
	}
	return
}
func (packet *PacketPlayKeepAlive) Write(player *Player) (err error) {
	if player.protocol >= V1_12_2 {
		err = player.WriteUInt64(uint64(packet.id))
	} else if player.protocol <= V1_7_6 {
		err = player.WriteUInt32(uint32(packet.id))
	} else {
		err = player.WriteVarInt(packet.id)
	}
	if err != nil {
		log.Print(err)
		return
	}
	return
}
func (packet *PacketPlayKeepAlive) Handle(player *Player) {
	if player.protocol > V1_8 {
		if player.keepalive != packet.id {
			player.Kick("Invalid keepalive")
		}
	} else {
		player.keepalive = packet.id
	}
	player.keepalive = 0
}
func (packet *PacketPlayKeepAlive) Id() int {
	return 0x1F
}

type PacketPlayJoinGame struct {
	entity_id     uint32
	gamemode      Gamemode
	dimension     Dimension
	difficulty    Difficulty
	max_players   uint8
	level_type    LevelType
	reduced_debug bool
}

func (packet *PacketPlayJoinGame) Read(player *Player, length int) (err error) {
	return
}
func (packet *PacketPlayJoinGame) Write(player *Player) (err error) {
	if player.protocol <= V1_9 {
		err = player.WriteUInt8(uint8(packet.entity_id))
	} else {
		err = player.WriteUInt32(packet.entity_id)
	}
	if err != nil {
		log.Print(err)
		return
	}
	err = player.WriteUInt8(uint8(packet.gamemode))
	if err != nil {
		log.Print(err)
		return
	}
	err = player.WriteUInt32(uint32(packet.dimension))
	if err != nil {
		log.Print(err)
		return
	}
	err = player.WriteUInt8(uint8(packet.difficulty))
	if err != nil {
		log.Print(err)
		return
	}
	err = player.WriteUInt8(packet.max_players)
	if err != nil {
		log.Print(err)
		return
	}
	err = player.WriteString(string(packet.level_type))
	if err != nil {
		log.Print(err)
		return
	}
	if player.protocol > V1_7_6 {
		err = player.WriteBool(packet.reduced_debug)
		if err != nil {
			log.Print(err)
			return
		}
	}
	return
}
func (packet *PacketPlayJoinGame) Handle(player *Player) {}
func (packet *PacketPlayJoinGame) Id() int {
	return 0x23
}

type PacketPlayerPositionLook struct {
	x           float64
	y           float64
	z           float64
	yaw         float32
	pitch       float32
	flags       uint8
	teleport_id int
}

func (packet *PacketPlayerPositionLook) Read(player *Player, length int) (err error) {
	return
}
func (packet *PacketPlayerPositionLook) Write(player *Player) (err error) {
	err = player.WriteFloat64(packet.x)
	if err != nil {
		log.Print(err)
		return
	}
	err = player.WriteFloat64(packet.y)
	if err != nil {
		log.Print(err)
		return
	}
	err = player.WriteFloat64(packet.z)
	if err != nil {
		log.Print(err)
		return
	}
	err = player.WriteFloat32(packet.yaw)
	if err != nil {
		log.Print(err)
		return
	}
	err = player.WriteFloat32(packet.pitch)
	if err != nil {
		log.Print(err)
		return
	}
	err = player.WriteUInt8(packet.flags)
	if err != nil {
		log.Print(err)
		return
	}
	if player.protocol > V1_8 {
		err = player.WriteVarInt(packet.teleport_id)
		if err != nil {
			log.Print(err)
			return
		}
	}
	return
}
func (packet *PacketPlayerPositionLook) Handle(player *Player) {}
func (packet *PacketPlayerPositionLook) Id() int {
	return 0x2E
}

type PacketPlayerListHeaderFooter struct {
	header *string
	footer *string
}

func (packet *PacketPlayerListHeaderFooter) Read(player *Player, length int) (err error) {
	return
}
func (packet *PacketPlayerListHeaderFooter) Write(player *Player) (err error) {
	var str string
	if packet.header == nil {
		str = `{"translate":""}`
	} else {
		str = *packet.header
	}
	err = player.WriteString(str)
	if err != nil {
		log.Print(err)
		return
	}
	if packet.footer == nil {
		str = `{"translate":""}`
	} else {
		str = *packet.footer
	}
	err = player.WriteString(str)
	if err != nil {
		log.Print(err)
		return
	}
	return
}
func (packet *PacketPlayerListHeaderFooter) Handle(player *Player) {}
func (packet *PacketPlayerListHeaderFooter) Id() int {
	return 0x47
}
