# Planning conversation (Claude chat)

My prompts from the planning conversation, translated from Portuguese and shortened to the part
that drove a decision (cuts marked with [...]). Operational questions (git commands, terminal
output, "what's next") are left out. Each prompt is followed by its outcome.

## 1. Initial request

> I received the following assignment from Sezzle:
>
> ```
> Objective
> Build a full-stack calculator application with a React frontend and a backend microservice.
> The frontend should consume the backend API to perform basic and advanced arithmetic
> operations. Focus on clean design, maintainable code, and testable architecture.
>
> Requirements
> Functional
> Operations:
> - Addition, Subtraction, Multiplication, Division
> - Optional: Exponentiation, Square Root, Percentage
> Frontend (React):
> - Intuitive UI for entering input and displaying results
> - Input validation and error handling
> - Responsive design (basic mobile support)
> Backend (REST API):
> - Expose endpoints for calculator operations
> - Validate input and handle edge cases (division by zero, invalid data)
> - Return results in JSON format
>
> Non-Functional
> - Clean, readable, and idiomatic code (frontend and backend)
> - Unit tests covering key functionality for both layers
> - Documentation: setup instructions, API usage, and design rationale
> - Optional: Dockerfile for full-stack deployment
>
> Constraints
> - Frontend: React (TypeScript preferred)
> - Backend: Go is perferred
>
> Deliverables
> 1. Git repository with frontend and backend code
> 2. README with setup instructions, API examples, and design decisions
> 3. Unit tests and coverage report
> 4. Optional: Dockerfile to run frontend + backend together
>
> Instructions
> 1. Use any AI tooling you would like
> 2. Spend ~2–4 hours on this assignment. Prioritize correctness, clarity, and
>    maintainability over extra features.
> 3. Push your solution to GitHub, GitLab, or another Git repository.
> 4. Share the repository link with us for evaluation.
> 5. Share any prompts that you used in your work
> 6. Make sure your README includes:
>    - Setup instructions
>    - How to run the frontend and backend
>    - Examples of API calls (if using REST)
>    - Design decisions or assumptions
> ```
>
> I want to plan it, from creating the git repo to versioning, with PRs, issue breakdown, etc.
> They also asked for the prompts used. Give me the execution plan, with the prompts I should send,
> the paths and the whole process I should follow before actually starting anything.

Outcome: a plan with repository setup, seven issues, CLAUDE.md conventions, one prompt per issue,
and a log of prompts kept in the repository.

## 2. UI direction

> I want to use a modern, well-regarded UI/UX option. I don't want the calculator design to be
> done in a simplistic way. The backend should be the basics that make sense, but I'm concerned
> about the design. [...]

Outcome: Tailwind + shadcn/ui, with Claude Code proposing two visual directions for me to choose from.

## 3. Reusing history entries

> The hover works, but reuse always modifies the first number. I'm thinking of setting it to the
> last clicked input. [...]

Outcome: rejected, since the target would be invisible state once focus moves to the history item.
Reuse keeps filling the first operand, like a calculator's ANS, and focus moves to the next input.

## 4. Bug found during manual testing

> [...] While testing, I found a bug: pressing Esc with no input focused doesn't clear the inputs.

> The same Esc bug applies to Enter. Improve the issue description.

Outcome: issue #13, with Enter keeping its native behavior on focused buttons such as history items.