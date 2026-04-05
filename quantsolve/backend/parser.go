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

// THE PRE-PROCESSOR: Expands simple multipliers attached to brackets
func expandBrackets(input string) string {
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
	input = strings.ToLower(input) // Variable standardization fix
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
			case ">": minVals[v] = val + 1
			case ">=": minVals[v] = val
			case "<": maxVals[v] = val - 1
			case "<=": maxVals[v] = val
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

	// Check for Rational (fractions with variables in denominator) e.g., /x
	matched, _ := regexp.MatchString(`\/[a-z]`, input)
	if matched {
		return "rational"
	}

	// Check for Quadratics (contains ^2)
	if strings.Contains(input, "^2") {
		return "quadratic"
	}

	// Check for other Polynomials (contains ^3, ^4, etc.)
	if strings.Contains(input, "^") {
		return "polynomial"
	}

	// Default assumption if no special characters are found
	return "linear"
}