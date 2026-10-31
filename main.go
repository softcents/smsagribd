package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

type Service struct {
	Name    string
	Dir     string
	Bin     string
	Port    string
	cmd     *exec.Cmd
	mu      sync.Mutex
	running bool
	stop    bool
}

var (
	baseDir  = "/www/wwwroot/smsagri.softcents.com"
	services = []*Service{
		{
			Name: "superman-api",
			Dir:  filepath.Join(baseDir, "superman-api"),
			Bin:  filepath.Join(baseDir, "superman-api", "superman-api"),
			Port: "8081",
		},
		{
			Name: "customer-api",
			Dir:  filepath.Join(baseDir, "customer-api"),
			Bin:  filepath.Join(baseDir, "customer-api", "customer-api"),
			Port: "8082",
		},
		{
			Name: "sp-api",
			Dir:  filepath.Join(baseDir, "sp-api"),
			Bin:  filepath.Join(baseDir, "sp-api", "sp-api"),
			Port: "8083",
		},
	}
	startTime = time.Now()
)

func streamOutput(prefix string, r io.Reader) {
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			os.Stdout.Write([]byte(fmt.Sprintf("[%s] %s", prefix, string(buf[:n]))))
		}
		if err != nil {
			break
		}
	}
}

func (s *Service) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		s.mu.Lock()
		if s.stop {
			s.mu.Unlock()
			return
		}
		s.mu.Unlock()

		log.Printf("[ORCHESTRATOR] Starting %s in %s...", s.Name, s.Dir)
		cmd := exec.Command(s.Bin)
		cmd.Dir = s.Dir
		cmd.Env = os.Environ()

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Printf("[ORCHESTRATOR] Error creating stdout pipe for %s: %v", s.Name, err)
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			log.Printf("[ORCHESTRATOR] Error creating stderr pipe for %s: %v", s.Name, err)
		}

		if err := cmd.Start(); err != nil {
			log.Printf("[ORCHESTRATOR] Failed to start %s: %v. Retrying in 5 seconds...", s.Name, err)
			time.Sleep(5 * time.Second)
			continue
		}

		s.mu.Lock()
		s.cmd = cmd
		s.running = true
		s.mu.Unlock()

		if stdout != nil {
			go streamOutput(s.Name, stdout)
		}
		if stderr != nil {
			go streamOutput(s.Name, stderr)
		}

		log.Printf("[ORCHESTRATOR] %s started with PID %d", s.Name, cmd.Process.Pid)

		// Wait for process to exit
		err = cmd.Wait()

		s.mu.Lock()
		s.running = false
		s.cmd = nil
		shouldStop := s.stop
		s.mu.Unlock()

		if shouldStop {
			log.Printf("[ORCHESTRATOR] %s stopped as requested.", s.Name)
			return
		}

		log.Printf("[ORCHESTRATOR] %s exited with error: %v. Restarting in 3 seconds...", s.Name, err)
		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
		}
	}
}

func (s *Service) Stop() {
	s.mu.Lock()
	s.stop = true
	cmd := s.cmd
	s.mu.Unlock()

	if cmd != nil && cmd.Process != nil {
		log.Printf("[ORCHESTRATOR] Sending SIGTERM to %s (PID %d)...", s.Name, cmd.Process.Pid)
		_ = cmd.Process.Signal(syscall.SIGTERM)

		// Wait up to 10 seconds, then kill if needed
		done := make(chan error, 1)
		go func() {
			_, err := cmd.Process.Wait()
			done <- err
		}()

		select {
		case <-done:
			log.Printf("[ORCHESTRATOR] %s stopped gracefully.", s.Name)
		case <-time.After(10 * time.Second):
			log.Printf("[ORCHESTRATOR] %s did not terminate in 10s, killing...", s.Name)
			_ = cmd.Process.Kill()
		}
	}
}

func main() {
	log.Println("[ORCHESTRATOR] SMS Whiteinfo BD / SMS Aggregator Master Orchestrator Starting...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	for _, s := range services {
		wg.Add(1)
		go s.Start(ctx, &wg)
	}

	// aaPanel Status Server on Port 8080
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		statusMap := make(map[string]any)
		allUp := true

		for _, s := range services {
			s.mu.Lock()
			running := s.running
			s.mu.Unlock()

			// Test health check
			healthOK := false
			resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%s/health", s.Port))
			if err == nil && resp.StatusCode == http.StatusOK {
				healthOK = true
				_ = resp.Body.Close()
			}

			statusMap[s.Name] = map[string]any{
				"port":     s.Port,
				"running":  running,
				"health":   healthOK,
			}

			if !running || !healthOK {
				allUp = false
			}
		}

		res := map[string]any{
			"status":    "running",
			"healthy":   allUp,
			"uptime":    time.Since(startTime).String(),
			"services":  statusMap,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		if !allUp {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		_ = json.NewEncoder(w).Encode(res)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","orchestrator":"healthy"}`))
	})

	server := &http.Server{
		Addr: "127.0.0.1:8080",
	}

	go func() {
		log.Println("[ORCHESTRATOR] aaPanel Status Server listening on 127.0.0.1:8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[ORCHESTRATOR] Status Server error: %v", err)
		}
	}()

	// Listen for termination signals (SIGINT, SIGTERM) - Ignore SIGHUP
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh
	log.Printf("[ORCHESTRATOR] Received signal %v. Shutting down microservices...", sig)
	cancel()

	// Shutdown HTTP status server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	_ = server.Shutdown(shutdownCtx)

	// Stop all child services in parallel
	var stopWg sync.WaitGroup
	for _, s := range services {
		stopWg.Add(1)
		go func(svc *Service) {
			defer stopWg.Done()
			svc.Stop()
		}(s)
	}
	stopWg.Wait()

	log.Println("[ORCHESTRATOR] All microservices stopped cleanly. Exiting.")
}
