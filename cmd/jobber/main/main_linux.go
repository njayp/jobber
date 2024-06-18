package main

import (
	"github.com/njayp/jobber/pkg/manager"
	"github.com/njayp/jobber/pkg/server"
	"github.com/spf13/cobra"
)

const url = ":9090"

func main() {
	rootCmd().Execute()
}

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.AddCommand(startCmd(), spawnCmd())
	return cmd
}

func startCmd() *cobra.Command {
	return &cobra.Command{
		Use: "start",
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := server.NewService()
			if err != nil {
				return err
			}
			return s.Serve(":9090")
		},
		Args: cobra.NoArgs,
	}
}

func spawnCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "spawn [flags] <id> <cmd> [<cmd args...>]",
		RunE: func(cmd *cobra.Command, args []string) error {
			return manager.Spawn(args[0], args[1], args[2:]...)
		},
		Args: cobra.MinimumNArgs(2),
	}
	return cmd
}
