package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
)

var (
	Port      int
	Listen    bool
	StaticDir string
)

var ServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve static files from a directory",
	Long:  `Starts a simple HTTP server to serve static files from a specified directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Verify the static directory exists
		if _, err := os.Stat(StaticDir); os.IsNotExist(err) {
			return fmt.Errorf("static directory '%s' does not exist", StaticDir)
		}

		// Determine host
		host := "127.0.0.1"
		if Listen {
			host = "0.0.0.0"
		}
		addr := fmt.Sprintf("%s:%d", host, Port)

		// Setup Router
		r := mux.NewRouter()

		// Serve static files
		// We use StripPrefix so the server doesn't look for /static/filename inside the folder
		// but rather serves the content of the folder at the root path.
		fs := http.FileServer(http.Dir(StaticDir))
		r.PathPrefix("/").Handler(fs)

		srv := &http.Server{
			Handler:      r,
			Addr:         addr,
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		}

		// Log absolute path for clarity
		absPath, _ := filepath.Abs(StaticDir)
		log.Printf("Serving %s on http://%s", absPath, addr)

		return srv.ListenAndServe()
	},
}

func init() {
	ServeCmd.Flags().IntVarP(&Port, "port", "p", 8080, "Port to listen on")
	ServeCmd.Flags().BoolVarP(&Listen, "listen", "l", false, "Toggle to listen on 0.0.0.0 instead of localhost")
	ServeCmd.Flags().StringVarP(&StaticDir, "dir", "d", "./static/", "Directory to serve static files from")
	RootCmd.AddCommand(ServeCmd)
}
