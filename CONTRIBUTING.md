# Project Conventions

## Coding Style Guidelines

- Follow Go conventions and use `go fmt`
- Use meaningful variable and function names
- Add comments for exported functions and complex logic
- Keep functions focused and reasonably sized
- Use structured logging with consistent format
- Honor golang's nature and prefer CapitalizedCamelCase for JSON/YAML codec
- Conform to [staticcheck](https://staticcheck.io/) to maintain high code quality

## Coding Best Practices

Follow general coding best practices, especially:

- DRY (Don't Repeat Yourself)
  - Leveraging existing code as much as possible. If necessary, revise the original code for more general usage.

## Branch Naming Convention

- Use format: `{author}/{descriptive-name}`
- Examples:
  - `john/add-new-gpx-command`
  - `sarah/fix-elevation-calculation`
  - `mike/improve-csv-parsing`
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
feat: add new GPX track simplification command

- Implement Douglas-Peucker algorithm for track simplification
- Add CLI flag for tolerance parameter
- Include progress bar for large files
- Add comprehensive test coverage

This new command allows users to reduce GPX file size while
maintaining track accuracy for visualization and analysis.

AI-Ratio: 0.3
```

```text
fix: resolve elevation data parsing issue

- Fix negative elevation values not being handled correctly
- Update elevation service error handling
- Add validation for coordinate bounds
- Improve error messages for debugging

Addresses issue where some GPX files with negative elevations
were causing parsing failures.

AI-Ratio: 0
```

```text
docs: update README with new command examples

- Add usage examples for all CLI commands
- Include sample GPX file processing workflows
- Document configuration options and environment variables
- Add troubleshooting section for common issues

Improves user experience by providing clear examples and
documentation for all available features.

AI-Ratio: 0.8
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

## Markdown Formatting Standards

- All markdown files must conform to markdownlint rules
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
