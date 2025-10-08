package websocket

import (
	"context"
	"encoding/json"
	"net/http"

	"fitness-trainer/internal/logger"
	desc "fitness-trainer/pkg/workouts"

	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	authHeader = "Authorization"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for now
		return true
	},
}

type ChatHandler struct {
	grpcConn *grpc.ClientConn
}

func NewChatHandler(grpcConn *grpc.ClientConn) *ChatHandler {
	return &ChatHandler{
		grpcConn: grpcConn,
	}
}

func (h *ChatHandler) HandleChat(w http.ResponseWriter, r *http.Request) {
	logger.Info("New WebSocket connection")

	// Get token from header
	token := r.Header.Get(authHeader)
	if token == "" {
		if initData := r.URL.Query().Get("init_data"); initData != "" {
			token = "tma " + initData
		}
	}

	if token == "" {
		logger.Error("No access token provided")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	// Create gRPC client
	client := desc.NewChatServiceClient(h.grpcConn)

	// Read single message from websocket
	messageType, message, err := conn.ReadMessage()
	if err != nil {
		logger.Error("Failed to read message: %v", err)
		return
	}

	if messageType != websocket.TextMessage {
		logger.Error("Invalid message type: %d", messageType)
		return
	}

	// Parse message as SendChatMessageRequest
	var req desc.SendChatMessageRequest
	if err := json.Unmarshal(message, &req); err != nil {
		logger.Error("Failed to unmarshal request: %v", err)
		h.sendError(conn, "Invalid request format")
		return
	}

	// Create context with metadata
	md := metadata.New(map[string]string{"authorization": token})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	// Call gRPC streaming method
	stream, err := client.SendChatMessageStream(ctx, &req)
	if err != nil {
		logger.Error("Failed to start stream: %v", err)
		h.sendError(conn, "Failed to start chat stream")
		return
	}

	// Read from stream and send to websocket
	for {
		resp, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				// Stream ended successfully
				break
			}
			logger.Error("Stream error: %v", err)
			h.sendError(conn, "Stream error")
			break
		}

		// Send response to websocket
		respBytes, err := protojson.Marshal(resp)
		if err != nil {
			logger.Error("Failed to marshal response: %v", err)
			h.sendError(conn, "Internal error")
			break
		}

		if err := conn.WriteMessage(websocket.TextMessage, respBytes); err != nil {
			logger.Error("Failed to write to websocket: %v", err)
			break
		}
	}

	conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

	// Close websocket after stream ends
	logger.Info("Closing WebSocket connection")
}

func (h *ChatHandler) sendError(conn *websocket.Conn, message string) {
	errorResp := map[string]string{"error": message}
	respBytes, _ := json.Marshal(errorResp)
	conn.WriteMessage(websocket.TextMessage, respBytes)
}
