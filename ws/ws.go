package ws

import (
	"encoding/json"
	"log"
	"time"

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

	peer_addr := noti_to_punch_req.ListenerPublicIP + ":" + noti_to_punch_req.ListenerPublicPort
	my_public_ip, my_public_port, err := handler.HandleNotifyToPunch(peer_addr)
	if err != nil {
		log.Println("Error while Handling NotifyToPunch:", err)
		log.Println("Source: handleNotifyToPunchRequest()")
		return
	}

	noti_to_punch_res := models.NotifyToPunchResponse{
		WorkspaceOwnerPublicIP:   my_public_ip,
		WorkspaceOwnerPublicPort: my_public_port,
		ListenerUsername:         noti_to_punch_req.ListenerUsername,
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

	PullWorkspace(noti_new_push.WorkspaceOwnerUsername, noti_new_push.WorkspaceName, conn)
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

		if msg.MessageType == "NotifyToPunchRequest" {
			log.Println("NotifyToPunchRequest Called")
			go handleNotifyToPunchRequest(conn, msg)
		} else if msg.MessageType == "NotifyNewPushToListeners" {
			log.Println("NotifyNewPushToListeners Called")
			go handleNotifyNewPushToListeners(msg, conn)
		} else if msg.MessageType == "RequestPunchFromReceiverResponse" {
			log.Println("RequestPunchFromReceiverResponse Called")
			go handleRequestPunchFromReceiverResponse(msg)
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
