package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Asset is shared across all files in the main package
type Asset struct {
	Name string
	Cost int
}

// THE DIVISION DISTRIBUTOR: Distributes division across terms inside a bracket
func distributeDivision(input string) string {
	re := regexp.MustCompile(`\(([^)]+)\)\/(\d*\.?\d+)`)

	return re.ReplaceAllStringFunc(input, func(match string) string {
		submatches := re.FindStringSubmatch(match)
		insideBrackets := submatches[1]
		divisorStr := submatches[2]

		divisor, _ := strconv.ParseFloat(divisorStr, 64)
		if divisor == 0 {
			return match
		} 

		assetRe := regexp.MustCompile(`([+-]?\d*\.?\d*)([a-z]?)`)

		expandedInside := assetRe.ReplaceAllStringFunc(insideBrackets, func(assetMatch string) string {
			assetSub := assetRe.FindStringSubmatch(assetMatch)
			costInsideStr := assetSub[1]
			variable := assetSub[2]

			if costInsideStr == "" && variable == "" {
				return ""
			}

			if costInsideStr == "" || costInsideStr == "+" {
				costInsideStr = "1"
			}
			if costInsideStr == "-" {
				costInsideStr = "-1"
			}

			costInside, _ := strconv.ParseFloat(costInsideStr, 64)
			newCost := costInside / divisor

			if newCost >= 0 {
				return fmt.Sprintf("+%.4f%s", newCost, variable)
			}
			return fmt.Sprintf("%.4f%s", newCost, variable)
		})

		return expandedInside
	})
}

// THE PRE-PROCESSOR: Expands simple multipliers attached to brackets
func expandBrackets(input string) string {
	impliedOneRe := regexp.MustCompile(`(^|[^0-9.])([a-z])`)
	input = impliedOneRe.ReplaceAllString(input, "${1}1${2}")

	re := regexp.MustCompile(`(\d*\.?\d+)\*?\(([^)]+)\)`)

	return re.ReplaceAllStringFunc(input, func(match string) string {
		submatches := re.FindStringSubmatch(match)
		multiplierStr := submatches[1]
		insideBrackets := submatches[2]

		multiplier, _ := strconv.ParseFloat(multiplierStr, 64)
		assetRe := regexp.MustCompile(`([+-]?\d*\.?\d+)([a-zA-Z])`)

		expandedInside := assetRe.ReplaceAllStringFunc(insideBrackets, func(assetMatch string) string {
			assetSub := assetRe.FindStringSubmatch(assetMatch)
			costInside, _ := strconv.ParseFloat(assetSub[1], 64)
			variable := assetSub[2]

			newCost := multiplier * costInside

			if newCost >= 0 {
				return fmt.Sprintf("+%.2f%s", newCost, variable)
			}
			return fmt.Sprintf("%.2f%s", newCost, variable)
		})

		return expandedInside
	})
}

// THE POLITE PARSER: Handles strings, brackets, and decimals
func parseEquation(input string) ([]Asset, int, error) {
	input = strings.ReplaceAll(input, " ", "")
	input = strings.ToLower(input)
	input = expandBrackets(input)

	parts := strings.Split(input, "=")
	if len(parts) != 2 {
		return nil, 0, fmt.Errorf("Oops! I couldn't find an '=' sign. Please add it so I know your target budget (e.g., = 100).")
	}

	rawTarget, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return nil, 0, fmt.Errorf("I had trouble reading the budget after the '=' sign ('%s'). Please make sure it is a valid number.", parts[1])
	}

	if rawTarget <= 0 {
		return nil, 0, fmt.Errorf("Your budget is currently set to %v. Please provide a budget greater than 0!", rawTarget)
	}

	targetInCents := int(rawTarget * 100)

	re := regexp.MustCompile(`(\d*\.?\d+)([a-zA-Z])`)
	matches := re.FindAllStringSubmatch(parts[0], -1)

	if len(matches) == 0 {
		return nil, 0, fmt.Errorf("I couldn't find any valid assets in your equation. Try attaching a cost to a letter, like '10x' or '5.25y'.")
	}

	var assets []Asset
	for _, match := range matches {
		rawCost, _ := strconv.ParseFloat(match[1], 64)
		costInCents := int(rawCost * 100)

		if costInCents <= 0 {
			return nil, 0, fmt.Errorf("Asset '%s' has a cost of 0. Free assets create infinite loops! Please give it a valid price.", match[2])
		}

		assets = append(assets, Asset{Name: match[2], Cost: costInCents})
	}
	return assets, targetInCents, nil
}

// THE RULES PARSER: Translates "x>5" into math limits
func parseConstraints(input string) (map[string]int, map[string]int) {
	minVals := make(map[string]int)
	maxVals := make(map[string]int)

	if input == "" {
		return minVals, maxVals
	}
	input = strings.ReplaceAll(input, " ", "")
	rules := strings.Split(input, ",")

	re := regexp.MustCompile(`([a-zA-Z]+)(>=|<=|>|<|=)(\d+)`)

	for _, rule := range rules {
		match := re.FindStringSubmatch(rule)
		if len(match) == 4 {
			v := match[1]
			op := match[2]
			val, _ := strconv.Atoi(match[3])

			switch op {
			case ">":
				minVals[v] = val + 1
			case ">=":
				minVals[v] = val
			case "<":
				maxVals[v] = val - 1
			case "<=":
				maxVals[v] = val
			case "=":
				minVals[v] = val
				maxVals[v] = val
			}
		}
	}
	return minVals, maxVals
}

// THE DETECTIVE: Scans the string to determine the equation family
func DetermineEquationType(input string) string {
	input = strings.ToLower(strings.ReplaceAll(input, " ", ""))

	// 1. Check for Fractions
	if matched, _ := regexp.MatchString(`\/[a-z]|\/\([a-z]`, input); matched {
		return "rational"
	}

	// 🌟 THE FIX: Check for High-Degree Exponents (^3, ^4, ^10, etc.) BEFORE checking for ^2!
	// Regex looks for ^ followed by 3-9, or ^ followed by any double-digit number
	if matched, _ := regexp.MatchString(`\^[3-9]|\^[1-9]\d+`, input); matched {
		return "polynomial"
	}

	// 2. Check for Quadratics
	if strings.Contains(input, "^2") {
		return "quadratic"
	}

	// Catch any weird edge case exponents
	if strings.Contains(input, "^") {
		return "polynomial"
	}

	re := regexp.MustCompile(`[a-z]`)
	matches := re.FindAllString(input, -1)
	uniqueVars := make(map[string]bool)
	for _, m := range matches {
		uniqueVars[m] = true
	}

	parts := strings.Split(input, "=")
	hasLetterOnRight := false
	if len(parts) == 2 {
		hasLetterOnRight, _ = regexp.MatchString(`[a-z]`, parts[1])
	}

	if len(uniqueVars) == 1 || hasLetterOnRight {
		return "polynomial"
	}

	return "linear"
}