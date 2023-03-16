package controllers

import (
	"encoding/json"
	"start/constants"
	socket_room "start/socket_room"
)

/*
Push socket khi user dc udp permission, để FE handle force logout
*/
func pushSocketUdpRole(roleId int64) {
	notiData := map[string]interface{}{
		"type":    constants.NOTIFICATION_ROLE_UDP,
		"role_id": roleId,
	}

	newFsConfigBytes, _ := json.Marshal(notiData)
	socket_room.Hub.Broadcast <- socket_room.Message{
		Data: newFsConfigBytes,
		Room: constants.NOTIFICATION_CHANNEL_ADMIN_1,
	}
}
