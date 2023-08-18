package main

import (
	"encoding/json"
	"fmt"
	"go.etcd.io/bbolt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

// TokenMappings is a map that holds the file containing
// the desired secret for each token
type TokenMappings map[string]string

func loadMappings(mappingsPath string) (TokenMappings, error) {
	var mappings TokenMappings

	// Read the JSON file
	mappingsFile, err := ioutil.ReadFile(mappingsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read mappings file: %w", err)
	}

	// Parse JSON into the mappings variable
	err = json.Unmarshal(mappingsFile, &mappings)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal token mappings: %w", err)
	}

	return mappings, nil
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: ./influxdb2-token-manipulator <influxd.bolt> <mappings.json>\n")
		os.Exit(1)
	}

	// Load token-to-path mappings from the JSON file
	tokenPaths, err := loadMappings(os.Args[2])
	if err != nil {
		fmt.Println("Error while loading token mappings:", err)
		os.Exit(1)
	}

	db, err := bbolt.Open(os.Args[1], 0666, nil)
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
	}
	defer db.Close()

	err = db.Update(func(tx *bbolt.Tx) error {
		authBucket := tx.Bucket([]byte("authorizationsv1"))
		if authBucket == nil {
			fmt.Println("Bucket 'authorizationsv1' not found.")
			os.Exit(1)
		}

		authIndex := tx.Bucket([]byte("authorizationindexv1"))
		if authIndex == nil {
			fmt.Println("Bucket 'authorizationindexv1' not found.")
			os.Exit(1)
		}

		return authBucket.ForEach(func(k, v []byte) error {
			var obj map[string]interface{}
			if err := json.Unmarshal(v, &obj); err != nil {
				fmt.Printf("Error unmarshalling JSON: %v\n", err)
				return nil // Continue processing other rows
			}

			description, ok := obj["description"].(string)
			if !ok {
				return nil // Skip if description is not present
			}

			identifierRegex := regexp.MustCompile(`[0-9a-f]{32}`)
			match := identifierRegex.FindString(description)
			if match == "" {
				return nil // Skip if description doesn't match regex
			}

			tokenPath, found := tokenPaths[match]
			if !found {
				return nil // Skip if match is not in lookup
			}
			delete(tokenPaths, match) // Remove entry from the map

			content, err := ioutil.ReadFile(tokenPath)
			if err != nil {
				fmt.Printf("Error reading new token file: %v\n", err)
				return nil // Continue processing other rows
			}
			newToken := strings.TrimSpace(string(content)) // Remove leading and trailing whitespace

			oldToken, ok := obj["token"].(string)
			if !ok {
				fmt.Printf("Skipping invalid token without .token\n")
				return nil // Skip if token is not present
			}

			if oldToken == newToken {
				return nil // Skip if token is already up-to-date
			}

			obj["token"] = newToken
			updatedValue, err := json.Marshal(obj)
			if err != nil {
				fmt.Printf("Error marshalling updated JSON: %v\n", err)
				return nil // Continue processing other rows
			}

			if err := authIndex.Delete([]byte(oldToken)); err != nil {
				fmt.Printf("Error deleting old token index in authorizationindexv1: %v\n", err)
				return nil // Continue processing other rows
			}

			if err := authIndex.Put([]byte(newToken), k); err != nil {
				fmt.Printf("Error adding new token index in authorizationindexv1: %v\n", err)
				return nil // Continue processing other rows
			}

			if err := authBucket.Put(k, updatedValue); err != nil {
				fmt.Printf("Error updating token in authorizationsv1: %v\n", err)
				return nil // Continue processing other rows
			}

			fmt.Printf("Updated token: '%s'\n", description)
			return nil
		})
	})
	if err != nil {
		fmt.Printf("Error during transaction: %v", err)
	}

	// Check if any tokens were not processed
	if len(tokenPaths) > 0 {
		fmt.Println("Warning: The following tokens were not encountered:")
		for token := range tokenPaths {
			fmt.Printf("- %s\n", token)
		}
	}
}
