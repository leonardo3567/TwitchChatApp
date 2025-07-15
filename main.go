package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// Message represents a single message from the database
type Message struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

func main() {
	// Open a connection to the PostgreSQL database
	db, err := sql.Open("postgres", "postgres://root:root@db:5432/test_db?sslmode=disable")
	if err != nil {
		fmt.Println("Error connecting to PostgreSQL database:", err)
		return
	}
	defer db.Close()

	// Create tables for each channel to store chat messages if not exists
	// You can modify the channel names and table names as needed
	channels := []string{"channel1", "channel2"} // Add more channel names if needed
	for _, channel := range channels {
		_, err := db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS messages_%s (
			id SERIAL PRIMARY KEY,
			username TEXT,
			message TEXT,
			userTimeStamp TIMESTAMP
		)`, channel))
		if err != nil {
			fmt.Println("Error creating table for channel", channel, ":", err)
			return
		}
	}

	// Define API endpoint handler to fetch messages
	http.HandleFunc("/api/messages", func(w http.ResponseWriter, r *http.Request) {
		// Enable CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Check the request method
		if r.Method == http.MethodOptions {
			return
		}

		// Get channel parameter from query string
		channel := r.URL.Query().Get("channel")
		if channel == "" {
			http.Error(w, "Channel parameter is required", http.StatusBadRequest)
			return
		}

		// Query messages from the database for the specified channel
		rows, err := db.Query(fmt.Sprintf("SELECT id, username, message, userTimeStamp FROM messages_%s ORDER BY id DESC LIMIT 10", channel))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Create a slice to store messages
		messages := make([]Message, 0)

		// Iterate over the rows and populate the messages slice
		for rows.Next() {
			var message Message
			err := rows.Scan(&message.ID, &message.Username, &message.Message, &message.Timestamp)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			messages = append(messages, message)
		}

		// Check for errors during iteration
		if err := rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Convert messages slice to JSON
		jsonBytes, err := json.Marshal(messages)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Set response headers and write JSON response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonBytes)
	})

	// Start the HTTP server
	fmt.Println("Server is running on :8080")
	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			fmt.Println("Error starting HTTP server:", err)
			return
		}
	}()

	// Prompt the user to input the channel to log messages
	fmt.Println("Enter the channel to log messages:")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		fmt.Println("Error reading input:", scanner.Err())
		return
	}
	channel := scanner.Text()

	// Twitch credentials
	oauth := "oauth:w6u9na8pejq46btedmwia86zadhzy9" // You can generate one from https://twitchapps.com/tmi/
	username := "gomes3567"

	// Connect to Twitch IRC server
	conn, err := net.Dial("tcp", "irc.chat.twitch.tv:6667")
	if err != nil {
		fmt.Println("Error connecting to Twitch IRC:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to Twitch IRC")

	// Authenticate with Twitch IRC server
	fmt.Fprintf(conn, "PASS %s\r\n", oauth)
	fmt.Fprintf(conn, "NICK %s\r\n", username)
	fmt.Fprintf(conn, "JOIN #%s\r\n", channel)

	// Create a reader to read messages from the Twitch IRC server
	reader := bufio.NewReader(conn)

	// Create a prepared statement for inserting messages into the database
	stmt, err := db.Prepare(fmt.Sprintf("INSERT INTO messages_%s (username, message, userTimeStamp) VALUES ($1, $2, $3)", channel))
	if err != nil {
		fmt.Println("Error preparing statement:", err)
		return
	}
	defer stmt.Close()

	// Continuously read messages from the Twitch IRC server
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading message:", err)
			return
		}

		// Print the message to the console
		fmt.Print(message)

		// Insert the message into the PostgreSQL database
		if strings.Contains(message, "PRIVMSG") {
			// Split the message by spaces to extract components
			parts := strings.Split(message, " ")
			// Extract the username from the message
			username := strings.Split(parts[0], "!")[0][1:]
			// Join the message parts starting from the fourth part
			messageText := strings.Join(parts[3:], " ")
			userTimeStamp := time.Now()
			// Insert the message into the database
			_, err := stmt.Exec(username, messageText, userTimeStamp)
			if err != nil {
				fmt.Println("Error inserting message into database:", err)
				return
			}
		}

		// Check if the message is a PING command from the server
		if strings.HasPrefix(message, "PING") {
			// Respond to the PING
		}
	}
}
