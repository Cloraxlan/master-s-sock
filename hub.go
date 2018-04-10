package sock

import ()

//Connection router that registers connections and handles read/write to them
type Hub struct {
	// Registered Clients.
	Clients map[*Client]bool

	// Inbound messages from the Clients.
	Messages map[int]string

	// Register requests from the Clients.
	register chan *Client
	// Last Message to Hub
	Input chan MessageBlock
	// Unregister requests from Clients.
	unregister chan *Client
}

// Initializes Hub Struct.
// Returns *Hub
func NewHub() *Hub {
	return &Hub{
		Input:      make(chan MessageBlock),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
		Messages:   make(map[int]string),
	}
}

//Initializes hub cycle
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.Clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}
			/*case message := <-h.broadcast:
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}*/
		}
	}
}
