package socket

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
	// "github.com/rs/cors"
)

func Allow(c *gin.Context) {

	if string(c.Request.Method) == http.MethodOptions {
		c.Request.Header.Add("Vary", "Origin")
		c.Request.Header.Add("Vary", "Access-Control-Allow-Methods")
		c.Request.Header.Add("Vary", "Access-Control-Allow-Headers")
		c.Request.Header.Set("Access-Control-Allow-Origin", "https://localhost:4000")
		c.Request.Header.Set("Access-Control-Allow-Methods", "*")
		c.Request.Header.Set("Access-Control-Allow-Headers", "*")
		return
	}

	respWriter := &respBodyWriter{body: &strings.Builder{}, ResponseWriter: c.Writer}
	c.Writer = respWriter
	respWriter.Header().Add("Vary", "Origin")
	respWriter.Header().Set("Access-Control-Allow-Origin", "https://localhost:4000")
	respWriter.Header().Set("Access-Control-Allow-Headers", "*")

	c.Next()
}

type respBodyWriter struct {
	gin.ResponseWriter
	body *strings.Builder
}

func (w respBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func RunSocket(port string) {
	router := gin.New()
	router.Use(Allow) // Để login từ localhost
	server := socketio.NewServer(nil)

	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		log.Println("connected:", s.ID())
		return nil
	})

	server.OnEvent("/", "notice", func(s socketio.Conn, msg string) {
		log.Println("notice:", msg)
		s.Emit("reply", "have "+msg)
	})

	server.OnEvent("/chat", "msg", func(s socketio.Conn, msg string) string {
		s.SetContext(msg)
		return "recv " + msg
	})

	server.OnEvent("/", "bye", func(s socketio.Conn) string {
		last := s.Context().(string)
		s.Emit("bye", last)
		s.Close()
		return last
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		log.Println("meet error:", e)
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		log.Println("closed", reason)
	})

	go func() {
		if err := server.Serve(); err != nil {
			log.Fatalf("socketio listen error: %s\n", err)
		}
	}()
	defer server.Close()

	router.GET("/socket.io/*any", gin.WrapH(server))
	router.POST("/socket.io/*any", gin.WrapH(server))

	if err := router.Run(port); err != nil {
		log.Fatal("failed run app: ", err)
	}
}
