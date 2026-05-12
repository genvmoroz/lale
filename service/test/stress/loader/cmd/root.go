package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/genvmoroz/lale/service/test/stress/loader/internal/core"
	createcard "github.com/genvmoroz/lale/service/test/stress/loader/internal/core/performer/create-card"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "loader",
	Run: func(cmd *cobra.Command, args []string) {
		if err := run(cmd); err != nil {
			cmd.PrintErrf("failed to run: %v", err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

const (
	laleServiceHostFlag       = "host"
	laleServicePortFlag       = "port"
	parallelUsersFlag         = "users"
	cardsPerUserFlag          = "cards-per-user"
	wordsPerCardFlag          = "words-per-card"
	createCardFlag            = "create-card"
	inspectCardFlag           = "inspect-card"
	promptCardFlag            = "prompt-card"
	getAllCardsFlag           = "get-all-cards"
	updateCardFlag            = "update-card"
	updateCardPerformanceFlag = "update-performance"
	getCardsToLearnFlag       = "get-to-learn"
	getCardsToRepeatFlag      = "get-to-repeat"
	getSentencesFlag          = "get-sentences"
	generateStoryFlag         = "generate-story"
	deleteCardFlag            = "delete-card"
)

func init() {
	rootCmd.Flags().String(laleServiceHostFlag, "", "Lale service host")
	rootCmd.Flags().Uint32(laleServicePortFlag, 0, "Lale service port")
	rootCmd.Flags().Uint32(parallelUsersFlag, 0, "Number of parallel users")
	rootCmd.Flags().Uint32(cardsPerUserFlag, 100, "Number of cards per user")
	rootCmd.Flags().Uint32(wordsPerCardFlag, 10, "Number of words per card")
	rootCmd.Flags().Bool(createCardFlag, false, "Enable create card action")
	rootCmd.Flags().Bool(inspectCardFlag, false, "Enable inspect card action")
	rootCmd.Flags().Bool(promptCardFlag, false, "Enable prompt card action")
	rootCmd.Flags().Bool(getAllCardsFlag, false, "Enable get all cards action")
	rootCmd.Flags().Bool(updateCardFlag, false, "Enable update card action")
	rootCmd.Flags().Bool(updateCardPerformanceFlag, false, "Enable update card performance action")
	rootCmd.Flags().Bool(getCardsToLearnFlag, false, "Enable get cards to learn action")
	rootCmd.Flags().Bool(getCardsToRepeatFlag, false, "Enable get cards to repeat action")
	rootCmd.Flags().Bool(getSentencesFlag, false, "Enable get sentences action")
	rootCmd.Flags().Bool(generateStoryFlag, false, "Enable generate story action")
	rootCmd.Flags().Bool(deleteCardFlag, false, "Enable delete card action")
}

func run(cmd *cobra.Command) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	builders := map[core.Action]core.PerformerBuilder{
		core.ActionCreateCard: createcard.NewBuilder(),
	}

	loader, err := core.NewLoader(builders)
	if err != nil {
		return fmt.Errorf("create loader: %w", err)
	}

	req, err := buildLoadRequest(cmd)
	if err != nil {
		return fmt.Errorf("build load request: %w", err)
	}

	if err = loader.Load(ctx, req); err != nil {
		return err
	}

	return nil
}

func buildLoadRequest(cmd *cobra.Command) (core.LoadRequest, error) {
	port, err := cmd.Flags().GetUint32(laleServicePortFlag)
	if err != nil {
		return core.LoadRequest{}, fmt.Errorf("parsing lale service port: %w", err)
	}
	parallelUsers, err := cmd.Flags().GetUint32(parallelUsersFlag)
	if err != nil {
		return core.LoadRequest{}, fmt.Errorf("parsing parallel users: %w", err)
	}
	cardsPerUser, err := cmd.Flags().GetUint32(cardsPerUserFlag)
	if err != nil {
		return core.LoadRequest{}, fmt.Errorf("parsing cards per user: %w", err)
	}
	wordsPerCard, err := cmd.Flags().GetUint32(wordsPerCardFlag)
	if err != nil {
		return core.LoadRequest{}, fmt.Errorf("parsing words per card: %w", err)
	}
	actionCreateCardEnabled, err := cmd.Flags().GetBool(createCardFlag)
	if err != nil {
		return core.LoadRequest{}, fmt.Errorf("parsing create card action: %w", err)
	}
	actionInspectCardEnabled, err := cmd.Flags().GetBool(inspectCardFlag)
	if err != nil {
		return core.LoadRequest{}, fmt.Errorf("parsing inspect card action: %w", err)
	}
	actionPromptCardEnabled, err := cmd.Flags().GetBool(promptCardFlag)
	if err != nil {
		return core.LoadRequest{}, fmt.Errorf("parsing prompt card action: %w", err)
	}
	actionGetAllCardsEnabled, err := cmd.Flags().GetBool(getAllCardsFlag)
	if err != nil {
		return core.LoadRequest{}, fmt.Errorf("parsing get all cards action: %w", err)
	}
	actionUpdateCardEnabled, err := cmd.Flags().GetBool(updateCardFlag)
	if err != nil {
		return core.LoadRequest{}, fmt.Errorf("parsing update card action: %w", err)
	}
	actionUpdateCardPerformanceEnabled, err := cmd.Flags().GetBool(updateCardPerformanceFlag)
	if err != nil {
		return core.LoadRequest{}, fmt.Errorf("parsing update card performance action: %w", err)
	}
	actionGetCardsToLearnEnabled, err := cmd.Flags().GetBool(getCardsToLearnFlag)
	if err != nil {
		return core.LoadRequest{}, fmt.Errorf("parsing get cards to learn action: %w", err)
	}
	actionGetCardsToRepeatEnabled, err := cmd.Flags().GetBool(getCardsToRepeatFlag)
	if err != nil {
		return core.LoadRequest{}, fmt.Errorf("parsing get cards to repeat action: %w", err)
	}
	actionGetSentencesEnabled, err := cmd.Flags().GetBool(getSentencesFlag)
	if err != nil {
		return core.LoadRequest{}, fmt.Errorf("parsing get sentences action: %w", err)
	}
	actionGenerateStoryEnabled, err := cmd.Flags().GetBool(generateStoryFlag)
	if err != nil {
		return core.LoadRequest{}, fmt.Errorf("parsing generate story action: %w", err)
	}
	actionDeleteCardEnabled, err := cmd.Flags().GetBool(deleteCardFlag)
	if err != nil {
		return core.LoadRequest{}, fmt.Errorf("parsing delete card action: %w", err)
	}

	return core.LoadRequest{
		LaleServiceHost:                    cmd.Flag(laleServiceHostFlag).Value.String(),
		LaleServicePort:                    port,
		ParallelUsers:                      parallelUsers,
		CardsPerUser:                       cardsPerUser,
		WordsPerCard:                       wordsPerCard,
		ActionCreateCardEnabled:            actionCreateCardEnabled,
		ActionInspectCardEnabled:           actionInspectCardEnabled,
		ActionPromptCardEnabled:            actionPromptCardEnabled,
		ActionGetAllCardsEnabled:           actionGetAllCardsEnabled,
		ActionUpdateCardEnabled:            actionUpdateCardEnabled,
		ActionUpdateCardPerformanceEnabled: actionUpdateCardPerformanceEnabled,
		ActionGetCardsToLearnEnabled:       actionGetCardsToLearnEnabled,
		ActionGetCardsToRepeatEnabled:      actionGetCardsToRepeatEnabled,
		ActionGetSentencesEnabled:          actionGetSentencesEnabled,
		ActionGenerateStoryEnabled:         actionGenerateStoryEnabled,
		ActionDeleteCardEnabled:            actionDeleteCardEnabled,
	}, nil
}
