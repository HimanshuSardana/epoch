package llm

import "fmt"

var CommitMessagePrompt = `You are a git commit message generator. Generate a concise commit message following the Conventional Commits format.

Allowed types: feat, fix, docs, style, refactor, perf, test, chore, build, ci, revert

Rules:
1. Start with type (e.g., feat:, fix:, etc.)
2. Use imperative mood (e.g., "add" not "added")
3. Keep subject line under 72 characters
4. Be specific about what changed

Example output: feat: add user authentication

Here is the git diff to analyze:
%s

Generate a commit message (just the message, no explanation):`

var ImproveCommitPrompt = `You are a git commit message improver. Improve the following commit message to follow Conventional Commits format.

Rules:
1. Start with type (feat, fix, docs, style, refactor, perf, test, chore, build, ci, revert)
2. Use imperative mood
3. Be specific and descriptive
4. Keep under 72 characters for subject

Current message:
%s

Improved message (just the message, no explanation):`

func BuildCommitMessagePrompt(diff string) string {
	return fmt.Sprintf(CommitMessagePrompt, diff)
}

func BuildImproveCommitPrompt(message string) string {
	return fmt.Sprintf(ImproveCommitPrompt, message)
}
