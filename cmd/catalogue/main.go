package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/build"
	"github.com/woolawin/catalogue/internal/target"
)

func main() {
	cmds := args()
	if err := cmds.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runAdd(cmd *cobra.Command, args []string) {
	internal.Add(args[0])
}

func runBuild(cmd *cobra.Command, args []string) {
	system, err := target.GetSystem()
	if err != nil {
		return
	}
	src, _ := cmd.Flags().GetString("src")
	dst, _ := cmd.Flags().GetString("dst")
	build.Build(build.BuildSrc(src), dst, system)
}

func args() *cobra.Command {
	var add = &cobra.Command{
		Use:   "add",
		Short: "Add a package to your system",
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

	var root = &cobra.Command{
		Use:   "catalogue",
		Short: "A simple CLI application built with Cobra",
		Long:  `mycli is a custom command-line tool. It provides features like starting a server with the 'serve' subcommand.`,
	}

	root.AddCommand(add)
	root.AddCommand(build)
	return root
}
