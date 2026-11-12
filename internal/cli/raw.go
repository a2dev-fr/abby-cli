package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

func newRawCmd() *cobra.Command {
	var bodyJSON string
	var queryPairs []string

	rawCmd := &cobra.Command{
		Use:   "raw <METHOD> <PATH>",
		Short: "Exécuter une requête HTTP brute authentifiée sur l'API Abby (ex: abby raw GET /v2/company/me)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			method := strings.ToUpper(strings.TrimSpace(args[0]))
			path := strings.TrimSpace(args[1])

			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}

			q := url.Values{}
			for _, kv := range queryPairs {
				parts := strings.SplitN(kv, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("paramètre --query invalide %q (attendu clé=valeur)", kv)
				}
				q.Add(parts[0], parts[1])
			}

			var body any
			if strings.TrimSpace(bodyJSON) != "" {
				var parsed any
				if err := json.Unmarshal([]byte(bodyJSON), &parsed); err != nil {
					return fmt.Errorf("--body doit être un JSON valide: %w", err)
				}
				body = parsed
			}

			var rawResp json.RawMessage
			if err := client.Do(context.Background(), method, path, q, body, &rawResp); err != nil {
				return err
			}
			if len(rawResp) == 0 {
				return printer.PrintJSON(map[string]any{"ok": true})
			}
			return printer.PrintJSON(rawResp)
		},
	}

	rawCmd.Flags().StringVarP(&bodyJSON, "body", "d", "", "Corps JSON de la requête (ex: '{\"title\":\"Nouveau titre\"}')")
	rawCmd.Flags().StringSliceVarP(&queryPairs, "query", "q", nil, "Paramètre(s) de query string sous la forme clé=valeur")
	return rawCmd
}
