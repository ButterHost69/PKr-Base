package ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/ButterHost69/PKr-Base/config"
	"github.com/ButterHost69/PKr-Base/handler"
	"github.com/ButterHost69/PKr-Base/models"

	"github.com/gorilla/websocket"
)

const (
	PONG_WAIT_TIME = 5 * time.Minute
	PING_WAIT_TIME = (PONG_WAIT_TIME * 9) / 10
)

var RequestPunchFromReceiverResponseMap = models.RequestPunchFromReceiverResponseMap{Map: map[string]models.RequestPunchFromReceiverResponse{}}

func handleNotifyToPunchRequest(conn *websocket.Conn, msg models.WSMessage) {
	msg_bytes, err := json.Marshal(msg.Message)
	if err != nil {
		log.Println("Error while marshaling:", err)
		log.Println("Source: handleNotifyToPunchRequest()")
		return
	}
	var noti_to_punch_req models.NotifyToPunchRequest
	if err := json.Unmarshal(msg_bytes, &noti_to_punch_req); err != nil {
		log.Println("Error while unmarshaling:", err)
		log.Println("Source: handleNotifyToPunchRequest()")
		return
	}
	log.Printf("Res: %#v", noti_to_punch_req)

	my_public_ip, my_public_port, my_private_ips, my_private_port, err := handler.HandleNotifyToPunchRequest(noti_to_punch_req.ListenerPublicIP, noti_to_punch_req.ListenerPublicPort, noti_to_punch_req.ListenerPrivateIPList, noti_to_punch_req.ListenerPrivatePort)
	if err != nil {
		log.Println("Error while Handling NotifyToPunch:", err)
		log.Println("Source: handleNotifyToPunchRequest()")
		return
	}

	noti_to_punch_res := models.NotifyToPunchResponse{
		WorkspaceOwnerPublicIP:      my_public_ip,
		WorkspaceOwnerPublicPort:    my_public_port,
		ListenerUsername:            noti_to_punch_req.ListenerUsername,
		WorkspaceOwnerPrivateIPList: my_private_ips,
		WorkspaceOwnerPrivatePort:   my_private_port,
	}

	res := models.WSMessage{
		MessageType: "NotifyToPunchResponse",
		Message:     noti_to_punch_res,
	}

	err = conn.WriteJSON(res)
	if err != nil {
		log.Println("Error while writing Response of Notify To Punch:", err)
		log.Println("Source: handleNotifyToPunchRequest()")
		return
	}
	log.Println("Response Sent to Server:", noti_to_punch_res)
	log.Println(noti_to_punch_res)
}

func handleNotifyNewPushToListeners(msg models.WSMessage, conn *websocket.Conn) {
	msg_bytes, err := json.Marshal(msg.Message)
	if err != nil {
		log.Println("Error while marshaling:", err)
		log.Println("Source: handleNotifyNewPushToListeners()")
		return
	}
	var noti_new_push models.NotifyNewPushToListeners
	if err := json.Unmarshal(msg_bytes, &noti_new_push); err != nil {
		log.Println("Error while unmarshaling:", err)
		log.Println("Source: handleNotifyNewPushToListeners()")
		return
	}
	log.Printf("Res: %#v", noti_new_push)

	err = PullWorkspace(noti_new_push.WorkspaceOwnerUsername, noti_new_push.WorkspaceName, conn)
	if err != nil {
		log.Println("Error while Pulling Data:", err)
		log.Println("Source: handleNotifyNewPushToListeners()")

		// Try Again only once after 5 minutes
		log.Println("Will Try Again after 5 minutes")
		time.Sleep(5 * time.Minute)
		err = PullWorkspace(noti_new_push.WorkspaceOwnerUsername, noti_new_push.WorkspaceName, conn)
		if err != nil {
			log.Println("Error while Pulling Data Again:", err)
			log.Println("Source: handleNotifyNewPushToListeners()")
		}
	}
}

func handleRequestPunchFromReceiverResponse(msg models.WSMessage) {
	msg_bytes, err := json.Marshal(msg.Message)
	if err != nil {
		log.Println("Error while marshaling:", err)
		log.Println("Source: handleNotifyToPunchResponse()")
		return
	}
	var msg_obj models.RequestPunchFromReceiverResponse
	if err := json.Unmarshal(msg_bytes, &msg_obj); err != nil {
		log.Println("Error while unmarshaling:", err)
		log.Println("Source: handleNotifyToPunchResponse()")
		return
	}
	RequestPunchFromReceiverResponseMap.Lock()
	RequestPunchFromReceiverResponseMap.Map[msg_obj.WorkspaceOwnerUsername] = msg_obj
	RequestPunchFromReceiverResponseMap.Unlock()
	log.Printf("Noti To Punch Res: %#v", msg_obj)
}

func handleWorkspaceOwnerIsOnline(msg models.WSMessage, conn *websocket.Conn) {
	msg_bytes, err := json.Marshal(msg.Message)
	if err != nil {
		log.Println("Error while marshaling:", err)
		log.Println("Source: handleWorkspaceOwnerIsOnline()")
		return
	}

	var msg_obj models.WorkspaceOwnerIsOnline
	if err := json.Unmarshal(msg_bytes, &msg_obj); err != nil {
		log.Println("Error while unmarshaling:", err)
		log.Println("Source: handleWorkspaceOwnerIsOnline()")
		return
	}

	user_config, err := config.ReadFromUserConfigFile()
	if err != nil {
		log.Println("Error while Reading User Config File:", err)
		log.Println("Source: handleWorkspaceOwnerIsOnline()")
		return
	}

	for _, server := range user_config.ServerLists {
		for _, workspace := range server.GetWorkspaces {
			if workspace.WorkspaceOwnerName == msg_obj.WorkspaceOwnerName {
				err := PullWorkspace(msg_obj.WorkspaceOwnerName, workspace.WorkspaceName, conn)
				if err != nil {
					log.Println("Error while Pulling Data:", err)
					log.Println("Source: handleWorkspaceOwnerIsOnline()")
				}
			}
		}
	}
}

func ReadJSONMessage(done chan struct{}, conn *websocket.Conn) {
	defer close(done)

	conn.SetReadDeadline(time.Now().Add(PONG_WAIT_TIME))
	conn.SetPongHandler(func(appData string) error {
		conn.SetReadDeadline(time.Now().Add(PONG_WAIT_TIME))
		return nil
	})

	for {
		var msg models.WSMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Read WebSocket Error from Server:", err)
			log.Println("Source: ReadJSONMessage()")
			return
		}

		log.Printf("Message: %#v", msg)

		switch msg.MessageType {
		case "NotifyToPunchRequest":
			log.Println("NotifyToPunchRequest Called")
			go handleNotifyToPunchRequest(conn, msg)
		case "NotifyNewPushToListeners":
			log.Println("NotifyNewPushToListeners Called")
			go handleNotifyNewPushToListeners(msg, conn)
		case "RequestPunchFromReceiverResponse":
			log.Println("RequestPunchFromReceiverResponse Called")
			go handleRequestPunchFromReceiverResponse(msg)
		case "WorkspaceOwnerIsOnline":
			log.Println("Workspace Owner is Online Called")
			go handleWorkspaceOwnerIsOnline(msg, conn)
		}
	}
}

func PingPongWriter(done chan struct{}, conn *websocket.Conn) {
	defer close(done)

	ticker := time.NewTicker(PING_WAIT_TIME)
	for {
		<-ticker.C
		if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			log.Println("No response of Ping from Server")
			return
		}
	}
}
