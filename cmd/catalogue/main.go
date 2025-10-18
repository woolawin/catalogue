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
	"github.com/woolawin/catalogue/internal/ext"
	"github.com/woolawin/catalogue/internal/pkge"
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
	src, err := cmd.Flags().GetString("src")
	if err != nil {
		fmt.Println("BAD COMMAND: --src flag missing")
		os.Exit(1)
	}
	dst, err := cmd.Flags().GetString("dst")
	if err != nil {
		fmt.Println("BAD COMMAND: --dst flag missing")
		os.Exit(1)
	}

	srcAbs, err := filepath.Abs(src)
	if err != nil {
		fmt.Println("BAD COMMAND: source is not a valid path")
		os.Exit(1)
	}

	api := ext.NewAPI(srcAbs)

	index, err := pkge.Build("index.catalogue.toml", api.Disk())
	if err != nil {
		fmt.Println("ERROR")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	err = build.Build(dst, index, system, api)
	if err != nil {
		fmt.Println("ERROR")
		fmt.Println(err.Error())
		os.Exit(1)
	} else {
		fmt.Println("COMPLETED")
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

	var root = &cobra.Command{
		Use:   "catalogue",
		Short: "The missing piece to APT. An APT Repository Middleware",
	}

	root.AddCommand(add)
	root.AddCommand(build)
	root.AddCommand(printSystem)
	root.AddCommand(version)
	return root
}
