package main

import (
	"fmt"
	"os"

	"github.com/njayp/jobber/pkg/client"
	"github.com/njayp/jobber/pkg/pb"
	"github.com/spf13/cobra"
)

const url = ":9090"

func main() {
	rootCmd().Execute()
}

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.AddCommand(startCmd(), stopCmd(), statusCmd(), streamCmd())
	return cmd
}

func startCmd() *cobra.Command {
	return &cobra.Command{
		Use: "start [flags] <cmd> [<cmd args...>]",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, err := client.NewJobberClient(url)
			if err != nil {
				return err
			}
			id, err := cli.Start(cmd.Context(), &pb.StartRequest{CmdString: args})
			if err != nil {
				return err
			}

			fmt.Printf("Job started with id: %s\n", id.Id)
			return nil
		},
		Args: cobra.MinimumNArgs(1),
	}
}

func stopCmd() *cobra.Command {
	return &cobra.Command{
		Use: "stop [flags] <id>",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, err := client.NewJobberClient(url)
			if err != nil {
				return err
			}
			_, err = cli.Stop(cmd.Context(), &pb.StopRequest{Id: args[0]})
			if err != nil {
				return err
			}

			fmt.Println("Job stopped.")
			return nil
		},
		Args: cobra.ExactArgs(1),
	}
}

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use: "status [flags] <id>",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, err := client.NewJobberClient(url)
			if err != nil {
				return err
			}
			resp, err := cli.Status(cmd.Context(), &pb.StatusRequest{Id: args[0]})
			if err != nil {
				return err
			}

			fmt.Printf("Job status: %s.", resp.State.String())
			return nil
		},
		Args: cobra.ExactArgs(1),
	}
}

type streamFlags struct {
	stderr bool
}

func streamCmd() *cobra.Command {
	sf := &streamFlags{}
	cmd := &cobra.Command{
		Use:  "stream",
		RunE: sf.run,
		Args: cobra.ExactArgs(1),
	}
	flags := cmd.Flags()
	flags.BoolVar(&sf.stderr, "stderr", false, "Stream stderr instead of stdout.")
	return cmd
}

func (s *streamFlags) run(cmd *cobra.Command, args []string) error {
	ss := pb.StreamSelect_Stdout
	if s.stderr {
		ss = pb.StreamSelect_Stderr
	}

	cli, err := client.NewJobberClient(url)
	if err != nil {
		return err
	}

	stream, err := cli.Stream(cmd.Context(), &pb.StreamRequest{Id: args[0], StreamSelect: ss})
	if err != nil {
		return err
	}

	for {
		resp, err := stream.Recv()
		if err != nil {
			return err
		}

		_, err = os.Stdout.Write(resp.Data)
		if err != nil {
			return err
		}
	}
}
