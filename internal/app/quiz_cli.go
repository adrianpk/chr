package app

import (
	"context"
	"fmt"
	"io"

	"github.com/urfave/cli/v3"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

func (factory commandFactory) quizCommand() *cli.Command {
	return &cli.Command{
		Name:  "quiz",
		Usage: "Manage runtime quizzes",
		Commands: []*cli.Command{
			factory.quizAnswerCommand(),
			factory.quizCancelCommand(),
			factory.quizCompleteCommand(),
			factory.quizListCommand(),
			factory.quizModeCommand(),
			factory.quizResumeCommand(),
			factory.quizSkipCommand(),
			factory.quizStartCommand(),
		},
	}
}

func (factory commandFactory) quizListCommand() *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "List runtime quizzes",
		Flags:  ticketProjectFlags(),
		Action: factory.runQuizList,
	}
}

func (factory commandFactory) quizStartCommand() *cli.Command {
	return &cli.Command{
		Name:   "start",
		Usage:  "Start a runtime quiz",
		Flags:  quizStartFlags(),
		Action: factory.runQuizStart,
	}
}

func (factory commandFactory) quizResumeCommand() *cli.Command {
	return &cli.Command{
		Name:   "resume",
		Usage:  "Resume a runtime quiz",
		Flags:  workspaceScopeFlags(),
		Action: factory.runQuizResume,
	}
}

func (factory commandFactory) quizAnswerCommand() *cli.Command {
	return &cli.Command{
		Name:   "answer",
		Usage:  "Answer the active runtime quiz question",
		Flags:  quizAnswerFlags(),
		Action: factory.runQuizAnswer,
	}
}

func (factory commandFactory) quizSkipCommand() *cli.Command {
	return &cli.Command{
		Name:   "skip",
		Usage:  "Skip the active runtime quiz question",
		Flags:  workspaceScopeFlags(),
		Action: factory.runQuizSkip,
	}
}

func (factory commandFactory) quizCompleteCommand() *cli.Command {
	return &cli.Command{
		Name:   "complete",
		Usage:  "Complete a runtime quiz",
		Flags:  workspaceScopeFlags(),
		Action: factory.runQuizComplete,
	}
}

func (factory commandFactory) quizCancelCommand() *cli.Command {
	return &cli.Command{
		Name:   "cancel",
		Usage:  "Cancel a runtime quiz",
		Flags:  workspaceScopeFlags(),
		Action: factory.runQuizCancel,
	}
}

func (factory commandFactory) quizModeCommand() *cli.Command {
	return &cli.Command{
		Name:      "mode",
		Usage:     "Switch runtime quiz mode",
		ArgsUsage: "<deep|fast>",
		Flags:     workspaceScopeFlags(),
		Action:    factory.runQuizMode,
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
			Value: string(quizcore.ModeDeep),
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

func (factory commandFactory) runQuizList(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("quiz list does not accept positional arguments")
	}

	operations, err := factory.projectService(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.ListQuizzes()
	if err != nil {
		return err
	}

	for _, state := range result.States {
		_, err = fmt.Fprintf(factory.io.Stdout,
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

func (factory commandFactory) runQuizStart(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("quiz start does not accept positional arguments")
	}

	mode := quizcore.Mode(cmd.String("mode"))
	if !mode.Valid() {
		return fmt.Errorf("quiz mode is invalid: %s", mode)
	}

	operations, err := factory.projectService(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.StartQuiz(QuizStartRequest{Scope: workspaceScopeFromCommand(cmd), Mode: mode})
	if err != nil {
		return err
	}

	return printQuizOperationResult(factory.io.Stdout, result)
}

func (factory commandFactory) runQuizResume(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("quiz resume does not accept positional arguments")
	}

	operations, err := factory.projectService(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.ResumeQuiz(QuizResumeRequest{Scope: workspaceScopeFromCommand(cmd)})
	if err != nil {
		return err
	}

	return printQuizOperationResult(factory.io.Stdout, result)
}

func (factory commandFactory) runQuizAnswer(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("quiz answer does not accept positional arguments")
	}

	operations, err := factory.projectService(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.AnswerQuiz(QuizAnswerRequest{Scope: workspaceScopeFromCommand(cmd), Answer: cmd.String("answer")})
	if err != nil {
		return err
	}

	return printQuizOperationResult(factory.io.Stdout, result)
}

func (factory commandFactory) runQuizSkip(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("quiz skip does not accept positional arguments")
	}

	operations, err := factory.projectService(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.SkipQuiz(QuizSkipRequest{Scope: workspaceScopeFromCommand(cmd)})
	if err != nil {
		return err
	}

	return printQuizOperationResult(factory.io.Stdout, result)
}

func (factory commandFactory) runQuizComplete(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("quiz complete does not accept positional arguments")
	}

	operations, err := factory.projectService(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.CompleteQuiz(QuizCompleteRequest{Scope: workspaceScopeFromCommand(cmd)})
	if err != nil {
		return err
	}

	return printQuizOperationResult(factory.io.Stdout, result)
}

func (factory commandFactory) runQuizCancel(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return fmt.Errorf("quiz cancel does not accept positional arguments")
	}

	operations, err := factory.projectService(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.CancelQuiz(QuizCancelRequest{Scope: workspaceScopeFromCommand(cmd)})
	if err != nil {
		return err
	}

	return printQuizOperationResult(factory.io.Stdout, result)
}

func (factory commandFactory) runQuizMode(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("quiz mode requires exactly one mode argument")
	}

	mode := quizcore.Mode(cmd.Args().First())
	if !mode.Valid() {
		return fmt.Errorf("quiz mode is invalid: %s", mode)
	}

	operations, err := factory.projectService(ctx, cmd)
	if err != nil {
		return err
	}

	result, err := operations.SwitchQuizMode(QuizModeRequest{Scope: workspaceScopeFromCommand(cmd), Mode: mode})
	if err != nil {
		return err
	}

	return printQuizOperationResult(factory.io.Stdout, result)
}

func workspaceScopeFromCommand(cmd *cli.Command) quizcore.Scope {
	return quizcore.Scope{CaseSlug: cmd.String("use-case")}
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
