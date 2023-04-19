package controllers

import (
	"start/callservices"
	"start/constants"
	"start/controllers/request"
)

/*
Push socket khi user dc udp permission, để FE handle force logout
*/
func pushSocketUdpRole(roleId int64) {
	notiData := map[string]interface{}{
		"type":    constants.NOTIFICATION_ROLE_UDP,
		"role_id": roleId,
	}

	// push mess socket
	reqSocket := request.MessSocketBody{
		Data: notiData,
		Room: "",
	}

	go callservices.PushMessInSocket(reqSocket)

	// newFsConfigBytes, _ := json.Marshal(notiData)
	// socket_room.Hub.Broadcast <- socket_room.Message{
	// 	Data: newFsConfigBytes,
	// 	Room: "",
	// }
}
