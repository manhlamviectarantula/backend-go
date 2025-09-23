// VẾT TÍCH THẤT BẠI KHI DÙNG SOCKET BẰNG GOLANG
// NÊN ĐÃ PHẢI TẠO THÊM MỘT BE NODEJS MỚI DÙNG SOCKET ĐƯỢC
// CHƯA RÕ TẠI SAO KHÔNG THỂ broadcastOnlineUsers
package main

import (
	"fmt"
	"net/http"
	"sync"

	socketio "github.com/googollee/go-socket.io"
	"github.com/googollee/go-socket.io/engineio"
	"github.com/googollee/go-socket.io/engineio/transport"
	"github.com/googollee/go-socket.io/engineio/transport/polling"
	"github.com/googollee/go-socket.io/engineio/transport/websocket"
)

var allowOriginFunc = func(r *http.Request) bool {
	return true
}

var userSocketMap = make(map[string]string)
var mu sync.Mutex

func GetReceiverSocketId(userId string) string {
	mu.Lock()
	defer mu.Unlock()
	return userSocketMap[userId]
}

func InitSocket() *socketio.Server {
	server := socketio.NewServer(&engineio.Options{
		Transports: []transport.Transport{
			&websocket.Transport{
				CheckOrigin: allowOriginFunc,
			},
			&polling.Transport{
				CheckOrigin: allowOriginFunc,
			},
		},
	})

	// Khi client connect
	server.OnConnect("/", func(s socketio.Conn) error {
		fmt.Println("🔌 A user connected:", s.ID())

		u := s.URL()
		userId := u.Query().Get("userId")
		if userId != "" {
			mu.Lock()
			userSocketMap[userId] = s.ID()
			mu.Unlock()
		}

		fmt.Println("✅ userId:", userId)

		broadcastOnlineUsers(server)
		return nil
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		fmt.Println("❌ A user disconnected:", s.ID())

		mu.Lock()
		for uid, sid := range userSocketMap {
			if sid == s.ID() {
				delete(userSocketMap, uid)
				break
			}
		}
		mu.Unlock()

		broadcastOnlineUsers(server)
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		fmt.Println("🔥 socket error:", s.ID(), e)
	})

	return server
}

func broadcastOnlineUsers(server *socketio.Server) {
	mu.Lock()
	defer mu.Unlock()

	users := make([]string, 0, len(userSocketMap))
	for uid := range userSocketMap {
		users = append(users, uid)
	}

	fmt.Println("get onlineU")
	server.BroadcastToNamespace("/", "getOnlineUsers", users)
}
