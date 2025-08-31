package main

import (
	"context"
	"crypto/tls"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/expki/backend/pixel-protocol/config"
	"github.com/expki/backend/pixel-protocol/database"
	"github.com/expki/backend/pixel-protocol/logger"
	"github.com/klauspost/compress/zstd"
	"github.com/quic-go/quic-go/http3"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
)

func main() {
	appCtx, stopApp := context.WithCancel(context.Background())

	// Load config
	var configPath string = "config.json"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}
	log.Default().Printf("Config path: %s\n", configPath)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Default().Printf("Creating sample config: %s\n", configPath)
		err = config.CreateSample(configPath)
		if err != nil {
			log.Fatalf("CreateSample: %v", err)
		}
	}
	log.Default().Println("Reading config...")
	configRaw, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("ReadFile %q: %v", configPath, err)
	}
	log.Default().Println("Parsing config...")
	cfg, err := config.ParseConfig(configRaw)
	if err != nil {
		log.Fatalf("ParseConfig: %v", err)
	}
	log.Default().Println("Loading TLS...")
	err = cfg.TLS.Configurate()
	if err != nil {
		log.Fatalf("Configurate: %v", err)
	}

	// Logger
	log.Default().Println("Setting log level:", cfg.LogLevel.String())
	logConf := zap.NewDevelopmentConfig()
	logConf.Level = cfg.LogLevel.Zap()
	l, err := logConf.Build()
	if err != nil {
		log.Fatalf("zap.NewDevelopment: %v", err)
	}
	logger.Initialize(l)
	defer l.Sync()

	// Database
	logger.Sugar().Info("Loading database...")
	db, err := database.New(appCtx, cfg.Database)
	if err != nil {
		logger.Sugar().Fatalf("database.New: %v", err)
	}

	// Server
	logger.Sugar().Info("Loading Server...")
	// TODO: servers

	// Create mux
	mux := http.NewServeMux()

	// HTTP
	server := http.Server{
		Handler: mux,
		Addr:    cfg.Server.HttpAddress,
	}

	// HTTP2
	server2 := http.Server{
		Handler: mux,
		Addr:    cfg.Server.HttpsAddress,
		TLSConfig: &tls.Config{
			GetCertificate: cfg.TLS.GetCertificate,
			ClientAuth:     tls.NoClientCert,
			NextProtos:     []string{"h2", "http/1.1"}, // Enable HTTP/2
		},
	}
	err = http2.ConfigureServer(&server2, &http2.Server{})
	if err != nil {
		logger.Sugar().Fatalf("http2.Server: %v", err)
	}

	// HTTP3 (QUIC)
	server3 := http3.Server{
		Handler: mux,
		Addr:    cfg.Server.Http3Address,
		TLSConfig: &tls.Config{
			GetCertificate: cfg.TLS.GetCertificate,
			ClientAuth:     tls.NoClientCert,
			NextProtos:     []string{"h3"}, // Enable HTTP/3
		},
		QUICConfig: nil, // Use default QUIC configuration
	}

	// Headers middleware
	middlewareHeaders := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Advertise HTTP/3 support via Alt-Svc header
			// Use the same host and port that the client connected to
			// Since HTTP/2 and HTTP/3 run on the same port (typically 443)
			if r.TLS != nil {
				// Only advertise HTTP/3 for HTTPS connections
				// Extract port from the Host header, default to 443 if not specified
				host := r.Host
				port := "443"
				if colonPos := strings.LastIndex(host, ":"); colonPos != -1 {
					port = host[colonPos+1:]
				}
				w.Header().Set("Alt-Svc", `h3=":`+port+`"; ma=86400`)
			}
			h.ServeHTTP(w, r)
		})
	}

	// Decompression middleware
	middlewareDecompression := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.Header.Get("Content-Encoding"), "zstd") {
				h.ServeHTTP(w, r)
				return
			}
			reader, err := zstd.NewReader(r.Body, zstd.WithDecoderLowmem(true))
			if err != nil {
				logger.Sugar().Errorf("Failed to create zstd reader: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			defer reader.Close()
			r.Body = &zstdRequestReader{ReadCloser: r.Body, Reader: reader}
			h.ServeHTTP(w, r)
		})
	}

	// Compression middleware
	middlewareCompression := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "zstd") {
				h.ServeHTTP(w, r)
				return
			}
			w.Header().Set("Content-Encoding", "zstd")
			encoder, err := zstd.NewWriter(w, zstd.WithEncoderLevel(zstd.SpeedFastest))
			if err != nil {
				logger.Sugar().Errorf("Failed to create zstd encoder: %v", err)
				h.ServeHTTP(w, r)
				return
			}
			defer encoder.Close()
			zstrw := &zstdResponseWriter{ResponseWriter: w, Writer: encoder}
			h.ServeHTTP(zstrw, r)
		})
	}

	// Routes: API
	mux.Handle("/api/upload", middlewareHeaders(middlewareDecompression(middlewareCompression(http.HandlerFunc(nil)))))

	// Routes: Static
	static := http.FileServerFS(distZstd)
	staticZstd := http.FileServerFS(distZstd)
	mux.Handle("/", middlewareHeaders(func() http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "zstd") {
				static.ServeHTTP(w, r)
				return
			}
			w.Header().Set("Content-Encoding", "zstd")
			staticZstd.ServeHTTP(w, r)
		})
	}()))

	// Start servers
	serverDone := make(chan struct{})
	go func() {
		logger.Sugar().Infof("HTTP server starting on %s", cfg.Server.HttpAddress)
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Sugar().Errorf("ListenAndServe http: %v", err)
		}
		close(serverDone)
	}()
	server2Done := make(chan struct{})
	go func() {
		logger.Sugar().Infof("HTTP2 server starting on %s", cfg.Server.HttpsAddress)
		err := server2.ListenAndServeTLS("", "")
		if err != nil && err != http.ErrServerClosed {
			logger.Sugar().Errorf("ListenAndServe https (http2): %v", err)
		}
		close(server2Done)
	}()
	server3Done := make(chan struct{})
	go func() {
		logger.Sugar().Infof("HTTP3 (QUIC) server starting on %s", cfg.Server.Http3Address)
		err := server3.ListenAndServe()
		if err != nil {
			logger.Sugar().Errorf("ListenAndServe http3: %v", err)
		}
		close(server3Done)
	}()

	// Interrupt signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Wait for servers to finish
	select {
	case <-interrupt:
		logger.Sugar().Info("Interrupt signal received")
	case <-appCtx.Done():
		logger.Sugar().Info("App stopped")
	case <-serverDone:
		logger.Sugar().Info("HTTP server stopped")
	case <-server2Done:
		logger.Sugar().Info("HTTP2 server stopped")
	case <-server3Done:
		logger.Sugar().Info("HTTP3 server stopped")
	}
	logger.Sugar().Info("Server shutting down")
	shutdownCtx, cancelShutdown := context.WithTimeout(appCtx, 3*time.Second)
	defer cancelShutdown()
	server.Shutdown(shutdownCtx)
	server2.Shutdown(shutdownCtx)
	server3.Close()
	server.Close()
	server2.Close()
	stopApp()
	db.Close()
	logger.Sugar().Info("Server stopped")
}

// zstdResponseWriter wraps the http.ResponseWriter to provide zstd compression
type zstdResponseWriter struct {
	http.ResponseWriter
	Writer *zstd.Encoder
}

func (w *zstdResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// zstdResponseWriter wraps the io.ReadClose to provide zstd decompression
type zstdRequestReader struct {
	io.ReadCloser
	Reader *zstd.Decoder
}

func (r *zstdRequestReader) Read(p []byte) (int, error) {
	return r.Reader.Read(p)
}
