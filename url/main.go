package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	neturl "net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	port     = "3155"
	addr     = ":" + port
	shortLen = 6
	baseURL  = "http://localhost:" + port + "/"
	cleanUp  = 10 * time.Minute
)

type Store interface {
	Save(key, target string) error
	Load(key string) (string, bool, error)
	Close() error
}

type memStore struct {
	mu   sync.RWMutex
	data map[string]string
}

func newMemStore() *memStore { return &memStore{data: make(map[string]string)} }

func (s *memStore) Save(key, target string) error {
	s.mu.Lock()
	s.data[key] = target
	s.mu.Unlock()
	return nil
}

func (s *memStore) Load(key string) (string, bool, error) {
	s.mu.RLock()
	target, ok := s.data[key]
	s.mu.RUnlock()
	return target, ok, nil
}

func (s *memStore) Close() error { return nil }

var base62 = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func generateID(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	out := make([]byte, n)
	for i := 0; i < n; i++ {
		out[i] = base62[int(b[i])%len(base62)]
	}
	return string(out), nil
}

type shortenReq struct {
	URL string `json:"url" binding:"required"`
}

func validateURL(u string) bool {
	parsed, err := neturl.ParseRequestURI(u)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		parsed2, err2 := neturl.ParseRequestURI("http://" + u)
		if err2 != nil {
			return false
		}
		return parsed2.Scheme != "" && parsed2.Host != ""
	}
	return parsed.Scheme != "" && parsed.Host != ""
}

func main() {
	store := newMemStore()
	defer store.Close()

	// optional background placeholder (keeps binary similar to prior)
	go func() {
		t := time.NewTicker(cleanUp)
		defer t.Stop()
		for range t.C {
			// placeholder
		}
	}()

	r := gin.Default()

	// don't "trust all proxies" warning for local dev
	_ = r.SetTrustedProxies([]string{})

	// root page (HTML form)
	r.GET("/", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, `
<!doctype html>
<html>
  <head><meta charset="utf-8"><title>URL Shortener</title></head>
  <body>
    <h3>Shorten a URL</h3>
    <form method="post" action="/shorten" id="f">
      <input type="text" name="url" placeholder="https://example.com" style="width: 420px"/>
      <button type="submit">Shorten</button>
    </form>
    <p>API: <code>POST /shorten {"url":"https://example.com"}</code></p>
  </body>
</html>
`)
	})

	// ignore favicon noise
	r.GET("/favicon.ico", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	// /shorten supports JSON (API) and form submits (HTML)
	r.POST("/shorten", func(c *gin.Context) {
		var u string

		// prefer JSON when Content-Type contains application/json
		if strings.Contains(strings.ToLower(c.GetHeader("Content-Type")), "application/json") {
			var req shortenReq
			if err := c.BindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
				return
			}
			u = strings.TrimSpace(req.URL)
		} else {
			// form submit or query param
			u = strings.TrimSpace(c.PostForm("url"))
			if u == "" {
				u = strings.TrimSpace(c.Query("url"))
			}
		}

		if !validateURL(u) {
			// if form submit respond HTML, else JSON
			if strings.Contains(strings.ToLower(c.GetHeader("Content-Type")), "application/json") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid url"})
			} else {
				c.String(http.StatusBadRequest, "Invalid URL\n")
			}
			return
		}

		// generate key with simple collision retry
		var key string
		for i := 0; i < 6; i++ {
			id, err := generateID(shortLen)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "id generation failed"})
				return
			}
			if _, ok, _ := store.Load(id); !ok {
				key = id
				break
			}
		}
		if key == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "id collision"})
			return
		}

		if err := store.Save(key, u); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "save failed"})
			return
		}

		short := baseURL + key
		// respond according to request type
		if strings.Contains(strings.ToLower(c.GetHeader("Content-Type")), "application/json") {
			c.JSON(http.StatusCreated, gin.H{"short_url": short})
		} else {
			// display simple page with result when coming from form
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(http.StatusCreated, fmt.Sprintf(`<p>Short URL: <a href="%s">%s</a></p><p><a href="/">Shorten another</a></p>`, short, short))
		}
	})

	// redirect
	r.GET("/:id", func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		if target, ok, _ := store.Load(id); ok {
			c.Redirect(http.StatusFound, target)
			return
		}
		c.AbortWithStatus(http.StatusNotFound)
	})

	log.Printf("listening on %s (baseURL=%s)", addr, baseURL)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
