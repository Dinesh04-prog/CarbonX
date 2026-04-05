package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Payload handles the JSON arriving from the React Rule Engine
type Payload struct {
	Equation    string `json:"equation"`
	Constraints string `json:"constraints"`
}

// THE WEBSOCKET CONTROLLER & TRAFFIC ROUTER
func handleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		var payload Payload
		err := conn.ReadJSON(&payload)
		if err != nil { break }

		// 1. Detect Equation Type
		eqType := DetermineEquationType(payload.Equation)

		// 2. Route the Data
		switch eqType {
		case "linear":
			// --- EXISTING LINEAR ALLOCATION LOGIC ---
			assets, target, err := parseEquation(payload.Equation)
			if err != nil {
				conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
				conn.WriteMessage(websocket.TextMessage, []byte("FINISHED"))
				continue
			}

			minVals, maxVals := parseConstraints(payload.Constraints)
			found := false
			checkCount := 0

			solveRecursive(assets, target, make(map[string]int), conn, &found, minVals, maxVals, &checkCount)

			if !found {
				conn.WriteMessage(websocket.TextMessage, []byte("No valid solutions exist within these constraints."))
			}

			conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("STATS:%d", checkCount)))
			
		case "quadratic":
			// --- EXISTING QUADRATIC LOGIC ---
			solveQuadratic(payload.Equation, conn)

		case "rational":
			// --- NEW RATIONAL LOGIC ---
			solveRational(payload.Equation, payload.Constraints, conn)
		case "polynomial":
			// --- NEW POLYNOMIAL LOGIC ---
			solvePolynomial(payload.Equation, conn)	

		default:
			// Catch-all for polynomials
			conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("ERROR: Engine currently does not support %s equations.", eqType)))
		}

		// Always close the stream when done
		conn.WriteMessage(websocket.TextMessage, []byte("FINISHED"))
	}
}

func main() {
	router := gin.Default()
	router.GET("/ws", handleWebSocket)
	fmt.Println("QuantSolve Master Engine running on http://localhost:8080")
	router.Run(":8080")
}