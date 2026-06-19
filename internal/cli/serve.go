package cli

import (
	"fmt"
	"net"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/statikowsky/mar/internal/version"
	"github.com/statikowsky/mar/internal/web"
)

func newServeCmd() *cobra.Command {
	var port int
	var noOpen bool
	c := &cobra.Command{
		Use:     "serve",
		Aliases: []string{"s"},
		Short:   "Serve the documentation UI locally",
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := openStore()
			if err != nil {
				return err
			}
			defer s.Close()

			ln, err := listenForServe("127.0.0.1", port, cmd.Flags().Changed("port"))
			if err != nil {
				return err
			}
			repo := repoName()
			projectPath := projectDir()
			url := "http://" + ln.Addr().String()
			fmt.Fprintf(cmd.OutOrStdout(), "mar %s serving %s at %s\n", version.Display(), repo, url)
			if !noOpen {
				openBrowser(url)
			}
			return web.NewServer(s, repo, projectPath).ServeListener(ln)
		},
	}
	c.Flags().IntVar(&port, "port", 7777, "port to listen on")
	c.Flags().BoolVar(&noOpen, "no-open", false, "do not open the browser")
	return c
}

func listenForServe(host string, port int, explicit bool) (net.Listener, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	ln, err := net.Listen("tcp", addr)
	if err == nil {
		return ln, nil
	}
	if explicit {
		return nil, fmt.Errorf("listen on %s: %w", addr, err)
	}
	ln, fallbackErr := net.Listen("tcp", host+":0")
	if fallbackErr != nil {
		return nil, fmt.Errorf("listen on %s:0: %w", host, fallbackErr)
	}
	return ln, nil
}

func repoName() string {
	wd, err := osGetwd()
	if err != nil {
		return "mar"
	}
	return filepath.Base(wd)
}

func projectDir() string {
	wd, err := osGetwd()
	if err != nil {
		return ""
	}
	return wd
}

func openBrowser(url string) {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd, args = "cmd", []string{"/c", "start"}
	default:
		cmd = "xdg-open"
	}
	exec.Command(cmd, append(args, url)...).Start()
}
