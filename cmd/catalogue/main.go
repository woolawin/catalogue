package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/build"
	"github.com/woolawin/catalogue/internal/clone"
	"github.com/woolawin/catalogue/internal/component"
	"github.com/woolawin/catalogue/internal/ext"
	"github.com/woolawin/catalogue/internal/target"
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

func runSystem(cmd *cobra.Command, args []string) {
	system, err := target.GetSystem()
	if err != nil {
		fmt.Println("ERROR")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	builder := strings.Builder{}
	builder.WriteString("Arhiteture: ")
	builder.WriteString(string(system.Architecture))
	builder.WriteString("\nOSReleaseID: ")
	builder.WriteString(system.OSReleaseID)
	builder.WriteString("\nOSReleaseVersion: ")
	builder.WriteString(system.OSReleaseVersion)
	builder.WriteString("\nOSReleaseVersionID: ")
	builder.WriteString(system.OSReleaseVersionID)
	builder.WriteString("\nOSReleaseVersionCodeName: ")
	builder.WriteString(system.OSReleaseVersionCodeName)

	fmt.Println(builder.String())
}

func runAdd(cmd *cobra.Command, args []string) {
	internal.Add(args[0])
}

func runBuild(cmd *cobra.Command, args []string) {
	system, err := target.GetSystem()
	if err != nil {
		fmt.Println("ERROR")
		fmt.Println(err.Error())
		os.Exit(1)
		return
	}
	src, _ := cmd.Flags().GetString("src")
	dst, _ := cmd.Flags().GetString("dst")

	srcAbs, err := filepath.Abs(src)
	if err != nil {
		fmt.Println("BAD COMMAND: source is not a valid path")
		os.Exit(1)
	}

	api := ext.NewAPI(srcAbs)

	config, err := component.Build("catalogue.toml", api.Disk())
	if err != nil {
		fmt.Println("ERROR")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	err = build.Build(dst, config, system, api)
	if err != nil {
		fmt.Println("ERROR")
		fmt.Println(err.Error())
		os.Exit(1)
	} else {
		fmt.Println("COMPLETED")
	}
}

func runClone(cmd *cobra.Command, args []string) {
	remote, _ := cmd.Flags().GetString("remote")
	local, _ := cmd.Flags().GetString("local")
	path, _ := cmd.Flags().GetString("path")

	var protocol clone.Protocol
	protocolCount := 0
	git, _ := cmd.Flags().GetBool("git")
	if git {
		protocolCount++
		protocol = clone.Git
	}

	if protocolCount != 1 {
		fmt.Println("ERROR: must specify --git")
		os.Exit(0)
	}

	api := ext.NewAPI("/")

	err := clone.Clone(protocol, remote, local, path, api)
	if err != nil {
		fmt.Println("ERROR")
		fmt.Println(err.Error())
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

	var build = &cobra.Command{
		Use:   "build",
		Short: "Build a package to be installed on your system",
		Long:  "",
		Run:   runBuild,
	}
	build.Flags().String("src", "", "Source directory to build from")
	build.Flags().String("dst", "", "Destination of the package archive")
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

	var root = &cobra.Command{
		Use:   "catalogue",
		Short: "The missing piece to APT. An APT Repository Middleware",
	}

	root.AddCommand(add)
	root.AddCommand(build)
	root.AddCommand(printSystem)
	root.AddCommand(version)
	root.AddCommand(clone)
	return root
}
