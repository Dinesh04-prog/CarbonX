package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}


type Asset struct {
	Name string
	Cost int
}

func parseEquation(input string) ([]Asset, int, error) {
	
	input = strings.ReplaceAll(input, " ", "")
	
	
	parts := strings.Split(input, "=")
	if len(parts) != 2 {
		return nil, 0, fmt.Errorf("invalid format: missing '='")
	}

	target, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, 0, fmt.Errorf("invalid target budget")
	}

	
	re := regexp.MustCompile(`(\d+)([a-zA-Z])`)
	matches := re.FindAllStringSubmatch(parts[0], -1)

	if len(matches) == 0 {
		return nil, 0, fmt.Errorf("could not detect any variables (e.g., 10x)")
	}

	var assets []Asset
	for _, match := range matches {
		cost, _ := strconv.Atoi(match[1])
		name := match[2]
		assets = append(assets, Asset{Name: name, Cost: cost})
	}

	return assets, target, nil
}


func solveRecursive(assets []Asset, target int, currentCombo map[string]int, conn *websocket.Conn, found *bool) {

	if len(assets) == 1 {
		lastAsset := assets[0]
		
		if target%lastAsset.Cost == 0 { 
			count := target / lastAsset.Cost
			currentCombo[lastAsset.Name] = count
			
	
			var parts []string
			for k, v := range currentCombo {
				parts = append(parts, fmt.Sprintf("%s=%d", k, v))
			}
			result := "Solution Found: " + strings.Join(parts, ", ")
			
	
			conn.WriteMessage(websocket.TextMessage, []byte(result))
			time.Sleep(20 * time.Millisecond) 
			*found = true
		}
		return
	}

	
	currentAsset := assets[0]
	maxAmount := target / currentAsset.Cost

	for i := 0; i <= maxAmount; i++ {
		
		newCombo := make(map[string]int)
		for k, v := range currentCombo {
			newCombo[k] = v
		}
		newCombo[currentAsset.Name] = i

		
		leftover := target - (i * currentAsset.Cost)
		solveRecursive(assets[1:], leftover, newCombo, conn, found)
	}
}

func solveAndStream(conn *websocket.Conn, assets []Asset, target int) {
	found := false
	initialCombo := make(map[string]int)
	
	
	solveRecursive(assets, target, initialCombo, conn, &found)

	if !found {
		conn.WriteMessage(websocket.TextMessage, []byte("No whole number solutions exist."))
	}
	conn.WriteMessage(websocket.TextMessage, []byte("FINISHED"))
}

func handleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil { return }
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil { break }

		assets, target, err := parseEquation(string(message))
		if err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte("Error: "+err.Error()))
			conn.WriteMessage(websocket.TextMessage, []byte("FINISHED"))
			continue
		}

		solveAndStream(conn, assets, target)
	}
}

func main() {
	router := gin.Default()
	router.GET("/ws", handleWebSocket)
	fmt.Println("QuantSolve Dynamic Engine running on http://localhost:8080")
	router.Run(":8080")
}