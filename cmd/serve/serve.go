package serve

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/nilptrderef/gogeo/frontend"
	"github.com/spf13/cobra"
)

var (
	Port    int
	Listen  bool
	DataDir string
)

var ServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve static files from a directory",
	Long:  `Starts a simple HTTP server to serve static files from a specified directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Verify the static directory exists
		if _, err := os.Stat(DataDir); os.IsNotExist(err) {
			return fmt.Errorf("data directory '%s' does not exist", DataDir)
		}

		// Determine host
		host := "127.0.0.1"
		if Listen {
			host = "0.0.0.0"
		}
		addr := fmt.Sprintf("%s:%d", host, Port)

		// Setup Router
		r := mux.NewRouter()
		r.HandleFunc("/data", Data)
		files, err := fs.Sub(frontend.Files, "build")
		if err != nil {
			return err
		}
		r.PathPrefix("/").Handler(http.FileServerFS(files))

		srv := &http.Server{
			Handler:      r,
			Addr:         addr,
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		}

		return srv.ListenAndServe()
	},
}

func init() {
	ServeCmd.Flags().IntVarP(&Port, "port", "p", 3000, "Port to listen on")
	ServeCmd.Flags().BoolVarP(&Listen, "listen", "l", false, "Toggle to listen on 0.0.0.0 instead of localhost")
	ServeCmd.Flags().StringVarP(&DataDir, "dir", "d", "./cmd/serve/static/", "Directory to serve static files from")
}

func Data(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open(filepath.Join(DataDir, "counties.msgpk"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "failed to load msgpk file"}`))
		return
	}
	defer file.Close()

	io.Copy(w, file)
}
