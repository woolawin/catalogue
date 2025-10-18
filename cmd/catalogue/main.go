package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/build"
	"github.com/woolawin/catalogue/internal/ext"
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
	src, err := cmd.Flags().GetString("src")
	if err != nil {
		fmt.Println("BAD COMMAND: --src flag missing")
	}
	dst, err := cmd.Flags().GetString("dst")
	if err != nil {
		fmt.Println("BAD COMMAND: --dst flag missing")
	}

	srcAbs, err := filepath.Abs(src)
	if err != nil {
		fmt.Println("BAD COMMAND: source is not a valid path")
	}

	api := ext.NewAPI(srcAbs)

	err = build.Build(dst, system, api)
	if err != nil {
		fmt.Println("ERROR")
		fmt.Println(err.Error())
	} else {
		fmt.Println("COMPLETED")
	}
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
