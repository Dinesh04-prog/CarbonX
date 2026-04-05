package main

import (
	"fmt"
	"strings"

	"github.com/gorilla/websocket"
)

// THE RATIONAL ENGINE: Flattens fractions into linear equations
func solveRational(equation string, constraints string, conn *websocket.Conn) {
	conn.WriteMessage(websocket.TextMessage, []byte("> Initiating Rational (Fraction) Parser..."))

	// 1. Sanitize
	equation = strings.ToLower(strings.ReplaceAll(equation, " ", ""))
	parts := strings.Split(equation, "=")

	if len(parts) != 2 {
		conn.WriteMessage(websocket.TextMessage, []byte("ERROR: Invalid format. Missing '='."))
		return
	}

	left := parts[0]
	right := parts[1]

	// 2. Find the fraction on the left side
	divParts := strings.Split(left, "/")
	if len(divParts) != 2 {
		conn.WriteMessage(websocket.TextMessage, []byte("ERROR: Currently only supporting simple single-fraction rational equations (e.g., 100/x = 5)."))
		return
	}

	numerator := divParts[0]
	denominator := divParts[1]
	
	// Clean up brackets if user typed 100/(x)
	denominator = strings.Trim(denominator, "()")

	conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("> Extracted Denominator: '%s'. Checking Division by Zero rules...", denominator)))

	// 3. The Flattener Trick: RightSide * (Denominator) = Numerator
	// This turns "100/x = 5" into "5*(x) = 100"
	flattenedEquation := fmt.Sprintf("%s*(%s)=%s", right, denominator, numerator)
	conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("> Flattening equation to: %s", flattenedEquation)))

	// 4. Send it through your existing Brain to expand the brackets
	expandedEquation := expandBrackets(flattenedEquation)
	conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("> Expanded to Linear Form: %s", expandedEquation)))

	// 5. Reroute directly to the Linear Solver!
	assets, target, err := parseEquation(expandedEquation)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("ERROR: Flattened equation failed linear parsing: " + err.Error()))
		return
	}

	minVals, maxVals := parseConstraints(constraints)
	found := false
	checkCount := 0

	solveRecursive(assets, target, make(map[string]int), conn, &found, minVals, maxVals, &checkCount)

	if !found {
		conn.WriteMessage(websocket.TextMessage, []byte("No valid solutions found for this fraction."))
	}

	// Output the search stats from the linear engine
	conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("STATS:%d", checkCount)))
}