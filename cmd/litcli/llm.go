package main

import (
	"fmt"
	"github.com/lightninglabs/lightning-terminal/litrpc"
	"github.com/urfave/cli"
	"strings"
)

var llmCommands = []cli.Command{
	{
		Name:     "llm",
		Usage:    "Interact with LLM-based analysis features.",
		Category: "LLM",
		Subcommands: []cli.Command{
			analyzeNodeCommand,
		},
	},
}

var analyzeNodeCommand = cli.Command{
	Name:     "analyze",
	Category: "LLM",
	Usage:    "Analyze the node's current state using LLM",
	Flags: []cli.Flag{
		cli.StringSliceFlag{
			Name:     "query",
			Usage:    "The natural language query to analyze the node data",
			Required: true,
		},
	},
	Action: func(c *cli.Context) error {
		query := strings.Join(c.StringSlice("query"), " ")
		if query == "" {
			return fmt.Errorf("query is required")
		}

		// Remove any surrounding quotes that might have been captured
		query = strings.Trim(query, `"'`)

		clientConn, cleanup, err := connectClient(c, true)
		if err != nil {
			return err
		}
		defer cleanup()

		fmt.Printf("Sending query: %q\n", query)

		client := litrpc.NewLLMClient(clientConn)
		ctx := getContext()

		req := &litrpc.AnalyzeNodeRequest{
			Query: query,
		}

		stream, err := client.AnalyzeNode(ctx, req)
		if err != nil {
			return err
		}

		// Process streaming responses
		for {
			resp, err := stream.Recv()
			if err != nil {
				// Check if we've reached the end of the stream
				if err.Error() == "EOF" {
					return nil
				}
				return err
			}

			// Print the analysis chunk
			fmt.Print(resp.Analysis)

			// If this is the final chunk, we're done
			if resp.Done {
				fmt.Println() // Add newline at the end
				return nil
			}
		}
	},
}
