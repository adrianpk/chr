package app

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chc/internal/core"
	"forge.adrianpk.com/chc/internal/webui"
	"forge.adrianpk.com/chc/internal/webui/ui"
)

func newWebCommand(stdout io.Writer, deps appDependencies) *cli.Command {
	return &cli.Command{
		Name:  "web",
		Usage: "Run the optional local Web UI",
		Commands: []*cli.Command{
			newWebServeCommand(stdout, deps),
		},
	}
}

func newWebServeCommand(stdout io.Writer, deps appDependencies) *cli.Command {
	return &cli.Command{
		Name:   "serve",
		Usage:  "Serve the optional local Web UI for the resolved project",
		Flags:  webServeFlags(),
		Action: webServeAction(stdout, deps),
	}
}

func webServeFlags() []cli.Flag {
	return append(projectRegistryFlags(),
		&cli.StringFlag{
			Name:  "dir",
			Usage: "project directory override",
			Value: ".",
		},
		&cli.StringFlag{
			Name:  "host",
			Usage: "web server bind host override",
		},
		&cli.IntFlag{
			Name:  "port",
			Usage: "web server bind port override",
		},
		&cli.BoolFlag{
			Name:  "open-browser",
			Usage: "open the local browser after the server starts",
		},
		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "print resolved web server config without listening on a socket",
		},
		&cli.StringFlag{
			Name:  "require-token",
			Usage: "web token mode override: auto, true, false, or always",
		},
	)
}

func webServeAction(stdout io.Writer, deps appDependencies) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runWebServe(ctx, cmd, stdout, deps)
	}
}

func runWebServe(ctx context.Context, cmd *cli.Command, stdout io.Writer, deps appDependencies) error {
	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, err := newCommandProjectContextResolver(env, deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
		ExplicitDir:           cmd.String("dir"),
		UseExplicitDir:        cmd.IsSet("dir"),
		AllowCurrentDirectory: true,
	})
	if err != nil {
		return err
	}

	config, err := core.LoadWorkspaceConfig(projectContext.ProjectDir)
	if err != nil {
		return err
	}

	webConfig := webConfigFromCommand(cmd, config.Web)
	if err := core.ValidateWebConfig(webConfig); err != nil {
		return err
	}

	resolved := resolvedWebServeConfig{
		ProjectDir:     projectContext.ProjectDir,
		Web:            webConfig,
		URL:            webURL(webConfig.Host, webConfig.Port),
		Env:            env,
		ProjectContext: projectContext,
		AssetsFS:       deps.assetsFS,
	}

	if cmd.Bool("dry-run") {
		return printWebServeDryRun(stdout, resolved)
	}

	if err := printWebServeStart(stdout, resolved); err != nil {
		return err
	}

	return serveWeb(ctx, resolved)
}

func webConfigFromCommand(cmd *cli.Command, config core.WebConfig) core.WebConfig {
	if cmd.IsSet("host") {
		config.Host = cmd.String("host")
	}

	if cmd.IsSet("port") {
		config.Port = cmd.Int("port")
	}

	if cmd.IsSet("open-browser") {
		config.OpenBrowser = cmd.Bool("open-browser")
	}

	if cmd.IsSet("require-token") {
		config.RequireToken = core.WebTokenMode(cmd.String("require-token"))
	}

	return config
}

type resolvedWebServeConfig struct {
	ProjectDir     string
	Web            core.WebConfig
	URL            string
	Env            core.UserEnvironment
	ProjectContext commandProjectContext
	AssetsFS       fs.FS
	Provider       webui.DataProvider
}

func printWebServeDryRun(stdout io.Writer, config resolvedWebServeConfig) error {
	if !core.WebHostIsLoopback(config.Web.Host) {
		if _, err := fmt.Fprintln(stdout, "warning: web ui is bound to a non-loopback interface"); err != nil {
			return err
		}
	}

	_, err := fmt.Fprintf(
		stdout,
		"web ui dry-run: %s\nproject: %s\nbind: %s\nport: %d\nenabled: %t\nopen_browser: %t\nrequire_token: %s\ntoken_env: %s\nstatic_assets: embedded\n",
		config.URL,
		config.ProjectDir,
		config.Web.Host,
		config.Web.Port,
		config.Web.Enabled,
		config.Web.OpenBrowser,
		config.Web.RequireToken,
		config.Web.TokenEnv,
	)

	return err
}

func printWebServeStart(stdout io.Writer, config resolvedWebServeConfig) error {
	if !core.WebHostIsLoopback(config.Web.Host) {
		if _, err := fmt.Fprintln(stdout, "warning: web ui is bound to a non-loopback interface"); err != nil {
			return err
		}
	}

	_, err := fmt.Fprintf(
		stdout,
		"web ui: %s\nproject: %s\nbind: %s\n",
		config.URL,
		config.ProjectDir,
		config.Web.Host,
	)

	return err
}

func serveWeb(ctx context.Context, config resolvedWebServeConfig) error {
	handler, err := newWebServeMux(config)
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:              net.JoinHostPort(config.Web.Host, strconv.Itoa(config.Web.Port)),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErr := make(chan error, 1)

	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			serverErr <- err

			return
		}

		serverErr <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return err
		}

		return ctx.Err()
	case err := <-serverErr:
		return err
	}
}

func newWebServeMux(config resolvedWebServeConfig) (http.Handler, error) {
	provider := config.Provider
	if provider == nil && config.Env != nil && config.AssetsFS != nil {
		provider = webViewProvider{
			env:            config.Env,
			assetsFS:       config.AssetsFS,
			projectContext: config.ProjectContext,
			config:         config,
		}
	}

	return webui.NewHandler(webui.HandlerOptions{
		Provider: provider,
		Data: webui.ShellData{
			Title:          "ChrysoCoord",
			ProjectDir:     config.ProjectDir,
			ProjectID:      projectID(config.ProjectDir),
			URL:            config.URL,
			Host:           config.Web.Host,
			Port:           config.Web.Port,
			RequireToken:   string(config.Web.RequireToken),
			RefreshSeconds: config.Web.RefreshIntervalSeconds,
			Navigation: []ui.NavLink{
				{Label: "Project", Href: "/", Active: true},
				{Label: "Board", Href: "/board"},
				{Label: "Dash", Href: "/dash"},
				{Label: "Cases", Href: "/cases"},
				{Label: "Dumps", Href: "/dumps"},
				{Label: "Quizzes", Href: "/quizzes"},
				{Label: "Specs", Href: "/specs"},
				{Label: "Plans", Href: "/plans"},
			},
		},
	})
}

func webURL(host string, port int) string {
	return "http://" + net.JoinHostPort(host, strconv.Itoa(port))
}
