package main

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

// THE COEFFCIENT EXTRACTOR: Pulls a, b, and c out of a string chunk
func extractCoefficients(side string, multiplier float64, a, b, c *float64) {
	if side == "" || side == "0" {
		return
	}

	// Force a leading sign so our regex always has a + or - to grab
	if !strings.HasPrefix(side, "+") && !strings.HasPrefix(side, "-") {
		side = "+" + side
	}

	// Regex explained: 
	// Group 1: The sign and number (e.g., "+3", "-", "+5.5")
	// Group 2: The variable part (e.g., "x^2", "x", or empty for constants)
	re := regexp.MustCompile(`([+-]\d*\.?\d*)(x\^2|x)?`)
	matches := re.FindAllStringSubmatch(side, -1)

	for _, match := range matches {
		coefStr := match[1]
		varType := match[2]

		// If regex catches a rogue sign with no number and no variable, skip it
		if (coefStr == "+" || coefStr == "-") && varType == "" {
			continue
		}

		// Handle implied 1s (e.g., "+x" -> "+1x", "-x^2" -> "-1x^2")
		if coefStr == "+" {
			coefStr = "+1"
		} else if coefStr == "-" {
			coefStr = "-1"
		}

		// Convert string to float
		coef, err := strconv.ParseFloat(coefStr, 64)
		if err != nil {
			continue
		}

		// Multiply by 1 for the left side, -1 for the right side (to move it across =)
		coef *= multiplier

		// Sort into a, b, or c buckets
		switch varType {
		case "x^2":
			*a += coef
		case "x":
			*b += coef
		default:
			*c += coef
		}
	}
}

// THE QUADRATIC ENGINE: Solves ax^2 + bx + c = 0
func solveQuadratic(equation string, conn *websocket.Conn) {
	conn.WriteMessage(websocket.TextMessage, []byte("> Initiating Quadratic Parsing Engine..."))

	// 1. Sanitize
	equation = strings.ToLower(strings.ReplaceAll(equation, " ", ""))

	// 2. Split by the equals sign
	parts := strings.Split(equation, "=")
	if len(parts) != 2 {
		conn.WriteMessage(websocket.TextMessage, []byte("ERROR: Invalid format. An equation must have exactly one '='."))
		return
	}

	var a, b, c float64

	// 3. Extract left side (multiplier +1) and right side (multiplier -1)
	extractCoefficients(parts[0], 1.0, &a, &b, &c)
	extractCoefficients(parts[1], -1.0, &a, &b, &c)

	conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("> Standard Form Achieved: %.2fx^2 + %.2fx + %.2f = 0", a, b, c)))

	// Failsafe: Did they type a linear equation by mistake? (e.g., 0x^2)
	if a == 0 {
		conn.WriteMessage(websocket.TextMessage, []byte("ERROR: The x^2 term cancelled out! This is actually a linear equation."))
		return
	}

	// 4. The Quadratic Formula
	discriminant := (b * b) - (4 * a * c)
	conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("> Discriminant calculated: %.2f", discriminant)))

if discriminant > 0 {
		root1 := (-b + math.Sqrt(discriminant)) / (2 * a)
		root2 := (-b - math.Sqrt(discriminant)) / (2 * a)
		// FIX: Changed "x=" to "x_1=" and "x_2=" so React doesn't overwrite the key!
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Solution Found: x_1=%.4f, x_2=%.4f", root1, root2)))
	} else if discriminant == 0 {
		root := -b / (2 * a)
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Solution Found: x=%.4f (Double Root)", root)))
	} else {
		conn.WriteMessage(websocket.TextMessage, []byte("ERROR: Discriminant is negative. No real solutions exist (requires imaginary numbers)."))
	}

	// Since we used a direct mathematical formula instead of brute-force checking possibilities, 
	// the search space is mathematically just "1" operation.
	conn.WriteMessage(websocket.TextMessage, []byte("STATS:1")) 
}