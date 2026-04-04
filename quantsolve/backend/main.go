package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)


var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}


func solveAndStream(conn *websocket.Conn, costX int, costY int, target int) {

	maxX := target / costX

	
	for x := 0; x <= maxX; x++ {
		leftover := target - (x * costX)

		
		if leftover%costY == 0 {
			y := leftover / costY
			
			
			result := fmt.Sprintf("Solution Found: x=%d, y=%d", x, y)
			
		
			conn.WriteMessage(websocket.TextMessage, []byte(result))
			
			
			time.Sleep(100 * time.Millisecond) 
		}
	}
	
	conn.WriteMessage(websocket.TextMessage, []byte("FINISHED"))
}

func handleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("Websocket error:", err)
		return
	}
	defer conn.Close()

	solveAndStream(conn, 10, 20, 100)
}

func main() {

	router := gin.Default()
	
	
	router.GET("/ws", handleWebSocket)

	fmt.Println("QuantSolve Engine running on http://localhost:8080")
	
	router.Run(":8080")
}