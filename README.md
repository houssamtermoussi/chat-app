# 🚀 Go Real-Time Chat (WebSocket)

A simple real-time chat application built with Go using WebSockets.

It allows multiple users to connect and exchange messages instantly without page refresh.

---

## ⚙️ Tech Stack

- Go (Backend)
- Gorilla WebSocket
- HTTP standard library

---

## 🧠 Architecture

Client (Browser)
    ↓ WebSocket
Go Server
    ↓
Hub (central message manager)
    ↓
Broadcast to all connected clients

The Hub is responsible for:
- Managing connected clients
- Broadcasting messages
- Handling user connections and disconnections safely

---

## 📂 Project Structure

main.go      → HTTP server + WebSocket route  
hub.go       → Hub logic (clients, broadcast, register/unregister)  
client.go    → Client read/write pumps  

---

## ⚡ Features

- Real-time messaging
- Multiple clients support
- WebSocket communication
- Automatic client cleanup on disconnect
- Broadcast system (send message to all users)

---

## 🚀 How to Run

1. Clone the repository

```bash
git clone https://github.com/houssamtermoussi/chat-app.git
