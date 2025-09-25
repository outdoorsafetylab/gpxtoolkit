# Project Conventions

## Coding Style Guidelines

- Follow Go conventions and use `go fmt`
- Use meaningful variable and function names
- Add comments for exported functions and complex logic
- Keep functions focused and reasonably sized
- Use structured logging with consistent format
- Honor golang's nature and prefer CapitalizedCamelCase for JSON/YAML codec
- Conform to [staticcheck](https://staticcheck.io/) to maintain high code quality
- Prefer pointers for function returns and struct members. e.g. `func GetItem() *Item` instead of `func GetItem() Item`, `struct { Member *Member }` instead of `struct { Member Member }`
- Avoid using `log.Fatalf(...)` except in cobra commands.

## Coding Best Practices

Follow general coding best practices, especially:

- DRY (Don't Repeat Yourself)
  - Leveraging existing code as much as possible. If necessary, revise the original code for more general usage.

## Branch Naming Convention

- Use format: `{author}/{descriptive-name}`
- Examples:
  - `john/add-new-metrics-endpoint`
- Use kebab-case for descriptive names (lowercase with hyphens)
- Keep descriptive names concise but clear
- The ticket ID is preferred but not required. And it must be an existing Jira ticket.

## Git Commit Message Convention

- Format:

  ```text
  {type}: {brief description in lowercase}

  AI-Ratio: {ai-ratio}
  ```

- Types:
  - `feat:` - New features
  - `fix:` - Bug fixes
  - `chore:` - Maintenance tasks, dependency updates, config changes
  - `docs:` - Documentation changes
  - `refactor:` - Code refactoring without feature changes
  - `test:` - Adding or updating tests
  - `ci:` - CI/CD pipeline changes
- AI-Ratio: a git commit trailer for annotating AI contribution.
  - `0` = fully human
  - `1` = fully AI-generated
  - in-between: partially AI-generated

The commits of merged PR should also follow this convention and append `(#PR)` in the end of title (the 1st line).

### Commit Message Structure

```text
{type}: {brief description in lowercase}

- Bullet point describing change 1
- Bullet point describing change 2
- Bullet point describing change 3

Optional longer explanation of the change and why it was made.

AI-Ratio: {ai-ratio}
```

### Examples

```text
chore: migrate github organization from abc to xyz

- Update GitHub workflows to use xyz organization
- Remove unused Repos configuration from config.yaml
- Keep PRCycleTime config for potential future use

This addresses the GitHub repository migration where SDKs moved
from the abc organization to xyz organization.

AI-Ratio: 0.5
```

```text
feat: add pull request cycle time analysis

- Implement PR cycle time calculation
- Add BigQuery integration for metrics storage
- Create CLI command for analysis report

This enables tracking of development velocity metrics.

AI-Ratio: 0
```

## Pull Request Guidelines

### Automatic PR Review Process

When asked to submit a pull request or a draft pull request, you must follow this workflow:

1. **Conduct Comprehensive PR Review First**
   - Perform a thorough review of all changes in the PR
   - Check code quality, style compliance, and best practices
   - Verify adherence to project conventions (commit messages, branch naming, etc.)
   - Review for potential bugs, security issues, or performance concerns
   - Ensure proper documentation and comments where needed
   - Validate that tests are included for new features or bug fixes

2. **Present Review Results**
   - Show a detailed review summary with categorized findings:
     - **Critical Issues**: Must be fixed before submission
     - **Recommendations**: Should be addressed for better code quality
     - **Minor Suggestions**: Optional improvements
     - **Positive Findings**: Good practices observed
   - For each finding, provide:
     - Clear description of the issue/suggestion
     - File location and line numbers
     - Recommended solution or improvement
     - Reasoning behind the recommendation

3. **User Confirmation Required**
   - If there are **Critical Issues**: Do NOT submit the PR automatically
     - Wait for user to address the issues and request review again
   - If there are **Recommendations** or **Minor Suggestions**:
     - Present the findings and ask user if they want to:
       - Address the recommendations first, OR
       - Proceed with submission as-is
   - If **No Issues Found**:
     - Present the positive review results
     - Ask for explicit confirmation before proceeding with submission

4. **PR Submission**
   - Only submit the PR after receiving user confirmation
   - Follow all project conventions for PR title, description, and formatting
   - Ensure PR template requirements are met

### Review Criteria

#### Code Quality Checks

- Code follows project style guidelines (Go conventions)
- Passes staticcheck linting with no errors
- Proper error handling and validation
- Meaningful variable and function names
- Appropriate use of design patterns
- No code duplication or overly complex logic

#### Go-Specific Checks

- Proper HTTP handler implementation
- Context usage for request handling
- Appropriate error handling and logging
- Template and static file handling
- Graceful handling of external API calls

#### Project Convention Compliance

- Branch naming follows `{author}/{descriptive-name}` format
- Commit messages follow the specified format and 72-character limit
- PR title follows `{type}: {description}` format
- All required sections in PR template are completed
- Proper Jira ticket linking

#### Documentation and Testing

- Public APIs have proper documentation (Go doc comments)
- Complex logic is well-commented
- Tests are included for new features or bug fixes
- README or other documentation updated if needed

### Example Review Output Format

```markdown
## PR Review Summary - Commit: [commit-hash-short]

### ✅ Positive Findings
- Good use of HTTP handler patterns
- Proper error handling implemented
- Code follows Go conventions
- Passes staticcheck linting

### ⚠️ Recommendations
- **File**: `router/session_handler.go:45`
  - **Issue**: Missing Go doc comment for exported function `ValidateAppID()`
  - **Suggestion**: Add documentation describing parameters and return value
  - **Reason**: Public APIs should be documented for better maintainability

### 💡 Minor Suggestions
- **File**: `router/session_handler.go:23`
  - **Issue**: Variable name could be more descriptive
  - **Suggestion**: Rename `data` to `sessionData` for clarity
  - **Reason**: Improves code readability

### 🎯 Overall Assessment
The code quality is good with no critical issues. The recommendations above would 
improve documentation and readability but are not blocking for submission.
```

This workflow ensures code quality while giving you control over when to submit the PR.

### Post-Submission PR Review

After a PR is successfully submitted, automatically conduct a follow-up review and post the results as a comment on the PR:

1. **Automatic Review After Submission**
   - Once the PR is created, immediately perform a comprehensive review of the entire PR content
   - Use the same review criteria as the pre-submission review
   - The commit hash in the title indicates when the review was conducted, not which specific commit was reviewed
   - Focus on any additional insights or observations after seeing the complete PR context

2. **Generate Review Comment**
   - Create a structured review comment using the same format as the `Example Review Output Format`
   - Include the latest commit hash in the title to indicate when the review was conducted
   - Review the entire PR content, not just the specific commit referenced in the title
   - Include all categories: Critical Issues, Recommendations, Minor Suggestions,
     and Positive Findings
   - Add any additional context that might be helpful for other reviewers

3. **Submit Review Comment**
   - Automatically post the review summary as a comment on the newly created PR
   - Use the GitHub review comment functionality to make it visible to all stakeholders
   - This provides immediate feedback and sets expectations for other reviewers

4. **Review Comment Format**
   - Use the exact same markdown format as shown in the Example Review Output Format
   - Include emojis and clear categorization for easy reading
   - Ensure all findings include file locations, specific issues, and reasoning

5. **Multiple Review Handling for Updated PRs**
   - Each review comment should include the latest commit hash in the title to indicate when the review was conducted
   - Always review the entire PR content, regardless of which commit hash is in the title
   - This allows tracking when reviews were conducted and which issues were addressed over time

This ensures that every PR gets an immediate, comprehensive review comment that helps maintain code quality and provides guidance for other team members reviewing the code.

## Markdown Formatting Standards

- All markdown files must conform to `markdownlint` rules
- Always run `markdownlint` after creating or updating markdown files
- Use consistent heading styles (prefer ATX-style: `# Heading`)
- Add space after hash marks in headings: `# Title` not `#Title`
- Use consistent list formatting with proper indentation
- No hard tabs - use spaces for indentation
- End files with a single newline
- Use reference-style links for better readability when appropriate
- Wrap lines at reasonable length (recommended: 80-100 characters)
- Use proper code block syntax with language specification
- Follow consistent emphasis formatting (prefer `**bold**` and `*italic*`)

### Markdown Examples

````markdown
# Main Title

## Section Title

### Subsection

- List item 1
- List item 2
  - Nested item
  - Another nested item

**Bold text** and *italic text* for emphasis.

Code blocks should specify language:

```go
func main() {
    fmt.Println("Hello, World!")
}
```

Reference links for better readability:
Check out the [markdownlint documentation][markdownlint] for more details.

[markdownlint]: https://github.com/DavidAnson/markdownlint
````

## For AI Agents

- **DO NOT START SERVER PROCESSES** Don't start the backend server or frontend development server by yourself. AI always got stuck by server processes and create zombie processes all the time. The worse of all, some IDEs like Cursor make the AI process read-only and prohibit human correction. So, please request the user to do it. The user probably using hot-reloading for other development, running or killing thread (with `pkill`) by yourself may interrupt the development or even cause ports are bound by unknown processes.
- **DO NOT COMMIT/PUSH GIT BY YOURSELF** Don't do git commit by yourself. Only do it upon clear and explicit request from user.
- **DO NOT STAGE CHANGES By YOURSELF** Don't add staged changes by yourself. If there is no staged changes to commit, double check with the user if you got a request to commit.
- **RUN TESTS AFTER MODIFICATION** Always run available unit tests after doing any changes.
