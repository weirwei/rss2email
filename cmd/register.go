package cmd

import (
	"regexp"
	"slices"

	"github.com/spf13/cobra"
	"github.com/weirwei/rss2email/constants"
	"github.com/weirwei/rss2email/service"
)

func init() {
	rootCmd.AddCommand(registerCmd)
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "db exec",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) <= 1 {
			cmd.Help()
			return
		}
		var (
			email         = args[0]
			subscriptions []constants.SubscriptionID
		)
		for _, v := range args[1:] {
			subscriptions = append(subscriptions, constants.SubscriptionID(v))
		}
		// subscription check
		for _, subscription := range subscriptions {
			if !slices.Contains(constants.AllSubscription, constants.SubscriptionID(subscription)) {
				cmd.Printf("subscription must be one of %v\nillegal subscription is [%s]", constants.AllSubscription, subscription)
				return
			}
		}
		// email check
		if !emailRegex.MatchString(email) {
			cmd.Printf("email %s is not valid\n", email)
			return
		}
		if err := service.Register(cmd.Context(), email, subscriptions); err != nil {
			cmd.Printf("register failed, %v\n", err)
			return
		}
		cmd.Printf("register success, email: %s, subscriptions: %v\n", email, subscriptions)
	},
}
