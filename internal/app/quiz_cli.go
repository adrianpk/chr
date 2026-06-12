package app

import (
	"context"
	"fmt"
	"io"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chc/internal/core"
)

func newQuizCommand(stdout io.Writer, deps appDependencies) *cli.Command {
	return &cli.Command{
		Name:  "quiz",
		Usage: "Manage runtime quizzes",
		Commands: []*cli.Command{
			newQuizAnswerCommand(stdout),
			newQuizCancelCommand(stdout),
			newQuizCompleteCommand(stdout),
			newQuizListCommand(stdout),
			newQuizModeCommand(stdout),
			newQuizResumeCommand(stdout),
			newQuizSkipCommand(stdout),
			newQuizStartCommand(stdout),
		},
	}
}

func newQuizListCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "List runtime quizzes",
		Flags:  ticketProjectFlags(),
		Action: quizListAction(stdout),
	}
}

func newQuizStartCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "start",
		Usage:  "Start a runtime quiz",
		Flags:  quizStartFlags(),
		Action: quizStartAction(stdout),
	}
}

func newQuizResumeCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "resume",
		Usage:  "Resume a runtime quiz",
		Flags:  workspaceScopeFlags(),
		Action: quizResumeAction(stdout),
	}
}

func newQuizAnswerCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "answer",
		Usage:  "Answer the active runtime quiz question",
		Flags:  quizAnswerFlags(),
		Action: quizAnswerAction(stdout),
	}
}

func newQuizSkipCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "skip",
		Usage:  "Skip the active runtime quiz question",
		Flags:  workspaceScopeFlags(),
		Action: quizSkipAction(stdout),
	}
}

func newQuizCompleteCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "complete",
		Usage:  "Complete a runtime quiz",
		Flags:  workspaceScopeFlags(),
		Action: quizCompleteAction(stdout),
	}
}

func newQuizCancelCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:   "cancel",
		Usage:  "Cancel a runtime quiz",
		Flags:  workspaceScopeFlags(),
		Action: quizCancelAction(stdout),
	}
}

func newQuizModeCommand(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:      "mode",
		Usage:     "Switch runtime quiz mode",
		ArgsUsage: "<deep|fast>",
		Flags:     workspaceScopeFlags(),
		Action:    quizModeAction(stdout),
	}
}

func workspaceScopeFlags() []cli.Flag {
	return append(ticketProjectFlags(),
		&cli.StringFlag{
			Name:  "use-case",
			Usage: "use-case slug",
		},
	)
}

func quizStartFlags() []cli.Flag {
	return append(workspaceScopeFlags(),
		&cli.StringFlag{
			Name:  "mode",
			Usage: "quiz mode: deep or fast",
			Value: string(core.QuizModeDeep),
		},
	)
}

func quizAnswerFlags() []cli.Flag {
	return append(workspaceScopeFlags(),
		&cli.StringFlag{
			Name:  "answer",
			Usage: "quiz answer",
		},
	)
}

func quizListAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runQuizList(ctx, cmd, stdout)
	}
}

func quizStartAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runQuizStart(ctx, cmd, stdout)
	}
}

func quizResumeAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runQuizResume(ctx, cmd, stdout)
	}
}

func quizAnswerAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runQuizAnswer(ctx, cmd, stdout)
	}
}

func quizSkipAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runQuizSkip(ctx, cmd, stdout)
	}
}

func quizCompleteAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runQuizComplete(ctx, cmd, stdout)
	}
}

func quizCancelAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runQuizCancel(ctx, cmd, stdout)
	}
}

func quizModeAction(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runQuizMode(ctx, cmd, stdout)
	}
}

func runQuizList(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("quiz list does not accept positional arguments")
	}

	operations, _, err := serviceFromCommand(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.ListQuizzes()
	if err != nil {
		return err
	}

	for _, state := range result.States {
		_, err = fmt.Fprintf(
			stdout,
			"%s\t%s\t%s\t%s\t%d/%d\n",
			state.ID,
			state.CaseSlug,
			state.Status,
			state.Mode,
			state.QuestionsAsked,
			state.QuestionBudget,
		)
		if err != nil {
			return fmt.Errorf("write quiz list result: %w", err)
		}
	}

	return nil
}

func runQuizStart(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("quiz start does not accept positional arguments")
	}

	mode := core.QuizMode(cmd.String("mode"))
	if !mode.Valid() {
		return fmt.Errorf("quiz mode is invalid: %s", mode)
	}

	operations, _, err := serviceFromCommand(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.StartQuiz(QuizStartRequest{Scope: workspaceScopeFromCommand(cmd), Mode: mode})
	if err != nil {
		return err
	}

	return printQuizOperationResult(stdout, result)
}

func runQuizResume(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("quiz resume does not accept positional arguments")
	}

	operations, _, err := serviceFromCommand(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.ResumeQuiz(QuizResumeRequest{Scope: workspaceScopeFromCommand(cmd)})
	if err != nil {
		return err
	}

	return printQuizOperationResult(stdout, result)
}

func runQuizAnswer(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("quiz answer does not accept positional arguments")
	}

	operations, _, err := serviceFromCommand(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.AnswerQuiz(QuizAnswerRequest{Scope: workspaceScopeFromCommand(cmd), Answer: cmd.String("answer")})
	if err != nil {
		return err
	}

	return printQuizOperationResult(stdout, result)
}

func runQuizSkip(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("quiz skip does not accept positional arguments")
	}

	operations, _, err := serviceFromCommand(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.SkipQuiz(QuizSkipRequest{Scope: workspaceScopeFromCommand(cmd)})
	if err != nil {
		return err
	}

	return printQuizOperationResult(stdout, result)
}

func runQuizComplete(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("quiz complete does not accept positional arguments")
	}

	operations, _, err := serviceFromCommand(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.CompleteQuiz(QuizCompleteRequest{Scope: workspaceScopeFromCommand(cmd)})
	if err != nil {
		return err
	}

	return printQuizOperationResult(stdout, result)
}

func runQuizCancel(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("quiz cancel does not accept positional arguments")
	}

	operations, _, err := serviceFromCommand(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.CancelQuiz(QuizCancelRequest{Scope: workspaceScopeFromCommand(cmd)})
	if err != nil {
		return err
	}

	return printQuizOperationResult(stdout, result)
}

func runQuizMode(ctx context.Context, cmd *cli.Command, stdout io.Writer) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("quiz mode requires exactly one mode argument")
	}

	mode := core.QuizMode(cmd.Args().First())
	if !mode.Valid() {
		return fmt.Errorf("quiz mode is invalid: %s", mode)
	}

	operations, _, err := serviceFromCommand(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.SwitchQuizMode(QuizModeRequest{Scope: workspaceScopeFromCommand(cmd), Mode: mode})
	if err != nil {
		return err
	}

	return printQuizOperationResult(stdout, result)
}

func workspaceScopeFromCommand(cmd *cli.Command) core.QuizScope {
	return core.QuizScope{CaseSlug: cmd.String("use-case")}
}

func printQuizOperationResult(stdout io.Writer, result QuizOperationResult) error {
	_, err := fmt.Fprintf(
		stdout,
		"%s\nquiz: %s %s %s %d/%d\n",
		result.Message,
		result.State.ID,
		result.State.CaseSlug,
		result.State.Status,
		result.State.QuestionsAsked,
		result.State.QuestionBudget,
	)
	if err != nil {
		return fmt.Errorf("write quiz operation result: %w", err)
	}

	if result.State.ActiveQuestion != nil {
		_, err = fmt.Fprintf(stdout, "question: %s\n", result.State.ActiveQuestion.Prompt)
		if err != nil {
			return fmt.Errorf("write quiz operation result: %w", err)
		}
	}

	return nil
}
