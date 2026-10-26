//nolint:wsl_v5 // Web runtime assembly keeps adjacent guarded setup steps together.
package app

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
	"forge.adrianpk.com/chr/internal/webui"
	"forge.adrianpk.com/chr/internal/webui/ui"
)

func (factory commandFactory) webCommand() *cli.Command {
	return &cli.Command{
		Name:  "web",
		Usage: "Run the local Web UI",
		Commands: []*cli.Command{
			factory.webServeCommand(),
			factory.webLoginCommand(),
		},
	}
}

func (factory commandFactory) webServeCommand() *cli.Command {
	return &cli.Command{
		Name:   "serve",
		Usage:  "Serve the local Web UI for the resolved project",
		Flags:  webServeFlags(),
		Action: factory.runWebServe,
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
		&cli.BoolFlag{
			Name:  "allow-lan",
			Usage: "allow direct token-free Web UI access on a trusted internal network (default: false)",
			Value: false,
		},
		&cli.StringFlag{
			Name:  "require-token",
			Usage: "web token mode override: auto, true, false, or always",
		},
		&cli.BoolFlag{
			Name:  "require-login",
			Usage: "refuse startup unless a private single-user login is configured",
		},
	)
}

func (factory commandFactory) runWebServe(ctx context.Context, cmd *cli.Command) error {
	env, err := newUserEnvironmentFromCommand(cmd)
	if err != nil {
		return err
	}

	projectContext, err := newCommandProjectContextResolver(env, factory.deps.assetsFS).Resolve(ctx, commandProjectContextOptions{
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

	webConfig, err := webConfigFromCommand(cmd, config.Web)
	if err != nil {
		return err
	}

	resolved := resolvedWebServeConfig{
		ProjectDir:     projectContext.ProjectDir,
		Web:            webConfig,
		URL:            webURL(webConfig.Host, webConfig.Port),
		Env:            env,
		ProjectContext: projectContext,
		AssetsFS:       factory.deps.assetsFS,
		LogRequests:    factory.io.Stdout,
	}

	resolved.Login, err = loadWebLogin(env, cmd.Bool("require-login"))
	if err != nil {
		return err
	}
	if resolved.Login != nil {
		resolved.Web.RequireToken = core.WebTokenModeFalse
	}

	if cmd.Bool("dry-run") {
		return printWebServeDryRun(factory.io.Stdout, resolved)
	}

	if err := printWebServeStart(factory.io.Stdout, resolved); err != nil {
		return err
	}

	return serveWeb(ctx, resolved)
}

func webConfigFromCommand(cmd *cli.Command, config core.WebConfig) (core.WebConfig, error) {
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

	if cmd.IsSet("allow-lan") {
		config.AllowLAN = cmd.Bool("allow-lan")
	}

	if err := core.ValidateWebConfig(config); err != nil {
		return core.WebConfig{}, err
	}

	return applyWebLANAccess(config, config.AllowLAN, cmd.IsSet("host"))
}

func applyWebLANAccess(config core.WebConfig, enabled, explicitHost bool) (core.WebConfig, error) {
	config.AllowLAN = enabled

	if !enabled {
		return config, nil
	}

	if config.RequireToken == core.WebTokenModeTrue || config.RequireToken == core.WebTokenModeAlways {
		return core.WebConfig{}, fmt.Errorf("--allow-lan conflicts with required web token protection; explicitly set --require-token=false or disable --allow-lan")
	}

	if !explicitHost && core.WebHostIsLoopback(config.Host) {
		config.Host = "0.0.0.0"
	}

	config.RequireToken = core.WebTokenModeFalse

	return config, nil
}

type resolvedWebServeConfig struct {
	guidedHost     guidedWorkActivator
	Login          *webui.LoginCredentials
	ProjectDir     string
	Web            core.WebConfig
	URL            string
	Env            core.UserEnvironment
	ProjectContext commandProjectContext
	AssetsFS       fs.FS
	Provider       webui.DataProvider
	LogRequests    io.Writer
}

func printWebServeDryRun(stdout io.Writer, config resolvedWebServeConfig) error {
	if err := printWebExposureWarnings(stdout, config.Web, config.Login != nil); err != nil {
		return err
	}

	allowedOrigins := webAllowedOriginsSummary(config.Web.AllowedOrigins)

	_, err := fmt.Fprintf(
		stdout,
		"web ui dry-run: %s\nproject: %s\nbind: %s\nport: %d\nenabled: %t\nopen_browser: %t\nallow_lan: %t\nrequire_token: %s\ntoken_required: %t\ntoken_env: %s\nallowed_origins: %s\nstatic_assets: embedded\n",
		config.URL,
		config.ProjectDir,
		config.Web.Host,
		config.Web.Port,
		config.Web.Enabled,
		config.Web.OpenBrowser,
		config.Web.AllowLAN,
		config.Web.RequireToken,
		webTokenRequired(config.Web),
		config.Web.TokenEnv,
		allowedOrigins,
	)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(stdout, "login_required: %t\n", config.Login != nil)

	return err
}

func printWebExposureWarnings(stdout io.Writer, config core.WebConfig, loginEnabled bool) error {
	if core.WebHostIsLoopback(config.Host) {
		return nil
	}

	if _, err := fmt.Fprintln(stdout, "warning: web ui is bound to a non-loopback interface"); err != nil {
		return err
	}

	if webTokenRequired(config) || loginEnabled {
		return nil
	}

	if _, err := fmt.Fprintln(stdout, "warning: web ui token protection is disabled on a non-loopback interface"); err != nil {
		return err
	}

	return nil
}

func printWebServeStart(stdout io.Writer, config resolvedWebServeConfig) error {
	if err := printWebExposureWarnings(stdout, config.Web, config.Login != nil); err != nil {
		return err
	}

	_, err := fmt.Fprintf(
		stdout,
		"web ui: %s\nproject: %s\nbind: %s\nallow_lan: %t\ntoken_required: %t\nallowed_origins: %s\n",
		config.URL,
		config.ProjectDir,
		config.Web.Host,
		config.Web.AllowLAN,
		webTokenRequired(config.Web),
		webAllowedOriginsSummary(config.Web.AllowedOrigins),
	)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(stdout, "login_required: %t\n", config.Login != nil)

	return err
}

func serveWeb(ctx context.Context, config resolvedWebServeConfig) error {
	host := newGuidedWorkHost(ctx)
	defer host.close()
	config.guidedHost = host
	config.ProjectContext = webServeProjectContext(config)
	config.Provider = webServeProvider(config)

	handler, err := newWebServeMux(config)
	if err != nil {
		return err
	}
	if provider, ok := config.Provider.(webViewProvider); ok {
		host.start(provider.service)
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
	projectContext := webServeProjectContext(config)
	provider := webServeProvider(config)

	accessConfig := config.Web
	if config.Login != nil {
		accessConfig.RequireToken = core.WebTokenModeFalse
	}

	access, err := webAccessPolicy(accessConfig)
	if err != nil {
		return nil, err
	}

	return webui.NewHandler(webui.HandlerOptions{
		Login:       config.Login,
		Provider:    provider,
		Access:      access,
		LogRequests: config.LogRequests,
		Data: webui.ShellData{
			Title:          "chr",
			ProjectDir:     config.ProjectDir,
			ProjectID:      projectContext.ProjectID,
			ProjectName:    projectContext.ProjectName,
			URL:            config.URL,
			Host:           config.Web.Host,
			Port:           config.Web.Port,
			RequireToken:   string(config.Web.RequireToken),
			RefreshSeconds: config.Web.RefreshIntervalSeconds,
			Navigation: []ui.NavLink{
				{Label: "Cases", Href: "/cases"},
				{Label: "Dumps", Href: "/dumps"},
				{Label: "Quizzes", Href: "/quizzes"},
				{Label: "Specs", Href: "/specs"},
				{Label: "Plans", Href: "/plans"},
				{Label: "Board", Href: "/board"},
				{Label: "Dash", Href: "/dash"},
			},
		},
	})
}

func webAccessPolicy(config core.WebConfig) (webui.AccessPolicy, error) {
	policy := webui.AccessPolicy{
		RequireToken:   webTokenRequired(config),
		Origin:         webURL(config.Host, config.Port),
		AllowedOrigins: append([]string(nil), config.AllowedOrigins...),
	}

	if !policy.RequireToken {
		return policy, nil
	}

	policy.Token = strings.TrimSpace(os.Getenv(config.TokenEnv))
	if policy.Token == "" {
		return webui.AccessPolicy{}, fmt.Errorf("web token is required but %s is not set", config.TokenEnv)
	}

	return policy, nil
}

func webTokenRequired(config core.WebConfig) bool {
	switch config.RequireToken {
	case core.WebTokenModeTrue, core.WebTokenModeAlways:
		return true
	case core.WebTokenModeFalse:
		return false
	case core.WebTokenModeAuto:
		return !core.WebHostIsLoopback(config.Host)
	default:
		return false
	}
}

func webAllowedOriginsSummary(origins []string) string {
	if len(origins) == 0 {
		return "same-origin"
	}

	return strings.Join(origins, ",")
}

func webURL(host string, port int) string {
	return "http://" + net.JoinHostPort(host, strconv.Itoa(port))
}
