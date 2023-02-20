package controllers

import (
	"encoding/json"
	"start/constants"
	"start/socket"
)

/*
 Push socket khi user dc udp permission, để FE handle forse logout
*/
func pushSocketUdpRole(roleId int64) {
	notiData := map[string]interface{}{
		"type":    constants.NOTIFICATION_ROLE_UDP,
		"role_id": roleId,
	}

	newFsConfigBytes, _ := json.Marshal(notiData)
	socket.GetHubSocket().Broadcast <- newFsConfigBytes
}
