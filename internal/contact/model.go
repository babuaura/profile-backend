package contact

import "time"

type Message struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    Budget    string    `json:"budget,omitempty"`
    Message   string    `json:"message"`
    Source    string    `json:"source"`
    Status    string    `json:"status"`
    CreatedAt time.Time `json:"createdAt"`
}

type CreateRequest struct { Name string `json:"name"`; Email string `json:"email"`; Budget string `json:"budget"`; Message string `json:"message"`; Source string `json:"source"` }
type StatusRequest struct { Status string `json:"status"` }
