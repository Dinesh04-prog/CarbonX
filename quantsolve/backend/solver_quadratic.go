package main

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

// THE COEFFCIENT EXTRACTOR: Upgraded to handle ANY letter (p, m, y, etc.)
func extractCoefficients(side string, multiplier float64, a, b, c *float64, varName *string) {
	if side == "" || side == "0" {
		return
	}

	if !strings.HasPrefix(side, "+") && !strings.HasPrefix(side, "-") {
		side = "+" + side
	}

	// NEW REGEX: Captures the number, ANY letter [a-z], and the ^2
	re := regexp.MustCompile(`([+-]\d*\.?\d*)([a-z])?(\^2)?`)
	matches := re.FindAllStringSubmatch(side, -1)

	for _, match := range matches {
		coefStr := match[1]
		letter := match[2]
		isSquared := match[3] == "^2"

		if letter != "" {
			*varName = letter // Capture the exact letter the user typed!
		}

		if (coefStr == "+" || coefStr == "-") && letter == "" {
			continue
		}

		if coefStr == "+" {
			coefStr = "+1"
		} else if coefStr == "-" {
			coefStr = "-1"
		}

		coef, err := strconv.ParseFloat(coefStr, 64)
		if err != nil {
			continue
		}

		coef *= multiplier

		// Sort into a, b, or c buckets based on the variable
		if letter != "" {
			if isSquared {
				*a += coef
			} else {
				*b += coef
			}
		} else {
			*c += coef
		}
	}
}

// THE QUADRATIC ENGINE: Solves ax^2 + bx + c = 0
func solveQuadratic(equation string, conn *websocket.Conn) {
	conn.WriteMessage(websocket.TextMessage, []byte("> Initiating Quadratic Parsing Engine..."))

	// 1. Sanitize
	equation = strings.ToLower(strings.ReplaceAll(equation, " ", ""))

	// 🌟 FIX: Add implied 1s so "-p" automatically becomes "-1p"
	impliedOneRe := regexp.MustCompile(`(^|[^0-9.])([a-z])`)
	equation = impliedOneRe.ReplaceAllString(equation, "${1}1${2}")

	// 2. Split by the equals sign
	parts := strings.Split(equation, "=")
	if len(parts) != 2 {
		conn.WriteMessage(websocket.TextMessage, []byte("ERROR: Invalid format. An equation must have exactly one '='."))
		return
	}

	var a, b, c float64
	varName := "x" // Default to x, but extractor will update it

	// 3. Extract left side (+1) and right side (-1)
	extractCoefficients(parts[0], 1.0, &a, &b, &c, &varName)
	extractCoefficients(parts[1], -1.0, &a, &b, &c, &varName)

	conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("> Standard Form Achieved: %.2f%s^2 + %.2f%s + %.2f = 0", a, varName, b, varName, c)))

	// Failsafe: Did they type a linear equation by mistake?
	if a == 0 {
		conn.WriteMessage(websocket.TextMessage, []byte("ERROR: The squared term cancelled out! This is actually a linear equation."))
		return
	}

	// 4. The Quadratic Formula
	discriminant := (b * b) - (4 * a * c)
	conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("> Discriminant calculated: %.2f", discriminant)))

	if discriminant > 0 {
		root1 := (-b + math.Sqrt(discriminant)) / (2 * a)
		root2 := (-b - math.Sqrt(discriminant)) / (2 * a)
		// Outputs the correct variable name (e.g., p_1 and p_2)
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Solution Found: %s_1=%.4f, %s_2=%.4f", varName, root1, varName, root2)))
	} else if discriminant == 0 {
		root := -b / (2 * a)
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Solution Found: %s=%.4f (Double Root)", varName, root)))
	} else {
		conn.WriteMessage(websocket.TextMessage, []byte("ERROR: Discriminant is negative. No real solutions exist (requires imaginary numbers)."))
	}

	conn.WriteMessage(websocket.TextMessage, []byte("STATS:1"))
}