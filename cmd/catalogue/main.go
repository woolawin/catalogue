package main

import (
	_ "embed"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/build"
	"github.com/woolawin/catalogue/internal/clone"
	"github.com/woolawin/catalogue/internal/config"
	"github.com/woolawin/catalogue/internal/daemon"
	"github.com/woolawin/catalogue/internal/ext"
)

//go:embed version.txt
var Version string

func main() {
	cmds := args()
	if err := cmds.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runVersion(cmd *cobra.Command, args []string) {
	fmt.Print(strings.TrimSpace(Version))
}

func runUpdate(cmd *cobra.Command, cliargs []string) {
	logger := internal.NewStdoutLogger(5)
	log := internal.NewLog(logger)
	log.Stage("cli")
	log.Info(7, "updating")

	client := daemon.NewClient(logger)
	component := ""
	if len(cliargs) != 0 {
		component = cliargs[0]
	}
	args := map[string]any{"component": component}
	ok, _, err := client.Send(daemon.Update, args)
	if err != nil || !ok {
		os.Exit(1)
	}
}

func runConfig(cmd *cobra.Command, args []string) {
	config, _ := ext.NewHost().GetConfig()
	fmt.Println("DefaultUser: ", config.DefaultUser)
}

func runSystem(cmd *cobra.Command, args []string) {
	system, err := ext.NewHost().GetSystem()
	if err != nil {
		fmt.Println("ERROR")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	builder := internal.Deb822{}
	builder.Add("Arhiteture", string(system.Architecture))
	builder.Add("OSReleaseID", system.OSReleaseID)
	builder.Add("OSReleaseVersion", system.OSReleaseVersion)
	builder.Add("OSReleaseVersionID", system.OSReleaseVersionID)
	builder.Add("OSReleaseVersionCodeName", system.OSReleaseVersionCodeName)

	fmt.Println(builder.String())
}

func runAdd(cmd *cobra.Command, cliargs []string) {
	logger := internal.NewStdoutLogger(5)
	log := internal.NewLog(logger)
	log.Stage("cli")
	protocol, remote, err := getProtocolAndRemote(cmd, cliargs)
	if err != nil {
		fmt.Println("ERROR")
		fmt.Println(err.Error())
		os.Exit(1)
	}
	log.Info(7, "adding component '%s'", remote)

	client := daemon.NewClient(logger)
	args := map[string]any{"protocol": protocol, "remote": remote}
	ok, _, err := client.Send(daemon.Add, args)
	if err != nil {
		fmt.Println("ERROR")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if !ok {
		fmt.Println("ERROR: could not add package")
		os.Exit(1)
	}

}

func runBuild(cmd *cobra.Command, args []string) {
	system, err := ext.NewHost().GetSystem()
	if err != nil {
		fmt.Println("ERROR")
		fmt.Println(err.Error())
		os.Exit(1)
		return
	}
	src, _ := cmd.Flags().GetString("src")
	dst, _ := cmd.Flags().GetString("dst")

	overrideSystem(&system, cmd)

	srcAbs, err := filepath.Abs(src)
	if err != nil {
		fmt.Println("BAD COMMAND: source is not a valid path")
		os.Exit(1)
	}

	dstAbs, err := filepath.Abs(dst)
	if err != nil {
		fmt.Println("BAD COMMAND: destination is not a valid path")
		os.Exit(1)
	}

	api := ext.NewAPI(srcAbs)
	dstFile, err := os.Create(dstAbs)
	if err != nil {
		fmt.Println("ERROR")
		fmt.Println(internal.ErrOf(err, "can not create .deb file '%s'", dstAbs))
		os.Exit(1)
		return
	}
	defer dstFile.Close()

	log := internal.NewLog(internal.NewStdoutLogger(8))

	ok := build.Build(dstFile, log, system, api)
	if !ok {
		os.Exit(1)
	} else {
		fmt.Println("COMPLETED")
	}
}

func runClone(cmd *cobra.Command, args []string) {
	remote, _ := cmd.Flags().GetString("remote")
	local, _ := cmd.Flags().GetString("local")
	path, _ := cmd.Flags().GetString("path")

	var protocol config.Protocol
	protocolCount := 0
	git, _ := cmd.Flags().GetBool("git")
	if git {
		protocolCount++
		protocol = config.Git
	}

	if protocolCount != 1 {
		fmt.Println("ERROR: must specify --git")
		os.Exit(0)
	}

	remoteURL, err := url.Parse(remote)
	if err != nil {
		fmt.Println("BAD COMMAND: remote is not a URL")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	api := ext.NewAPI("/")
	log := internal.NewLog(internal.NewStdoutLogger(9))
	opts := clone.NewOpts(
		config.Remote{Protocol: protocol, URL: remoteURL},
		local,
		path,
		nil,
	)
	ok := clone.Clone(opts, log, api)
	if !ok {
		os.Exit(1)
	}
}

func args() *cobra.Command {
	var add = &cobra.Command{
		Use:   "add",
		Short: "Add a package or repository to your local system (does not install)",
		Long:  ``,
		Run:   runAdd,
	}
	add.Flags().String("git", "", "Add from a git repository")

	var build = &cobra.Command{
		Use:   "build",
		Short: "Build a package to be installed on your system",
		Long:  "",
		Run:   runBuild,
	}
	build.Flags().String("src", "", "Source directory to build from")
	build.Flags().String("dst", "", "Destination of the package archive")
	build.Flags().String("architecture", "", "Architecture of package to build for")
	build.Flags().String("os-release-id", "", "OS Release ID of package to build for")
	build.Flags().String("os-release-version", "", "OS Release version of package to build for")
	build.Flags().String("os-release-version-id", "", "OS Release version ID of package to build for")
	build.Flags().String("os-release-version-code-name", "", "OS Release version code name of package to build for")
	build.MarkFlagRequired("src")
	build.MarkFlagRequired("dst")

	var printSystem = &cobra.Command{
		Use:   "system",
		Short: "Print system values used for targets",
		Long:  "",
		Run:   runSystem,
	}

	var version = &cobra.Command{
		Use:   "version",
		Short: "Print catalogue CLI tool version",
		Long:  "",
		Run:   runVersion,
	}

	var clone = &cobra.Command{
		Use:   "clone",
		Short: "Clone files from a remote source",
		Long:  "",
		Run:   runClone,
	}
	clone.Flags().String("remote", "", "The remote source to clone from")
	clone.Flags().String("local", "", "The local destination to clone to")
	clone.Flags().String("path", "", "The path to clone files from the remote")
	clone.Flags().Bool("git", false, "Clone via git")
	clone.MarkFlagRequired("remote")
	clone.MarkFlagRequired("local")
	clone.MarkFlagRequired("path")

	config := &cobra.Command{
		Use:   "config",
		Short: "Print catalogue configuration",
		Long:  "",
		Run:   runConfig,
	}

	update := &cobra.Command{
		Use:   "update",
		Short: "Update registry on for all or one component",
		Long:  "",
		Run:   runUpdate,
	}

	var root = &cobra.Command{
		Use:   "catalogue",
		Short: "The missing piece to APT. An APT Repository Middleware",
	}

	root.AddCommand(add)
	root.AddCommand(build)
	root.AddCommand(printSystem)
	root.AddCommand(version)
	root.AddCommand(clone)
	root.AddCommand(config)
	root.AddCommand(update)
	return root
}
