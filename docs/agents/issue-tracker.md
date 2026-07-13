# Issue Tracker: Linear

Issues, bugs, feature requests, specs, and ticket breakdowns for this repo live in Linear, not GitHub Issues.

## Linear Workspace

- Team: Dylans apps
- Team key: DYL
- Project: Shiet

Use the Linear MCP tools to create, read, list, update, and comment on issues.

## Conventions

- Create or update issues via Linear MCP, such as `save_issue` and `list_issues`.
- Do not create GitHub Issues unless the user explicitly asks.
- Use `gh` for pull requests, CI, checks, and repository operations only.
- When a skill says "publish to the issue tracker", create or update a Linear issue in the Shiet project.
- When a skill says "fetch the relevant ticket", fetch the matching Linear issue.

## Wayfinding operations

Used by `/wayfinder`. The **map** is a single Linear issue; **child** issues are tickets.

- **Map**: issue labelled `wayfinder:map`, holding Destination / Notes / Decisions so far / Not yet specified / Out of scope. Create or convert via `save_issue` with that label (Shiet project, Dylans apps team).
- **Child ticket**: issue with `parentId` set to the map, body starting with `## Question`, and label `wayfinder:<type>` (`research` / `prototype` / `grilling` / `task`). Claim by assigning the driving dev (`assignee: "me"`).
- **Blocking**: Linear native `blockedBy` / `blocks` relations on `save_issue`. A ticket is unblocked when every blocker is completed/canceled.
- **Frontier**: open children of the map with no open blocker and no assignee; take first in map/creation order.
- **Resolve**: post answer as a comment, mark the issue Done, append a one-line gist + link under the map’s Decisions so far.

## Pull Requests as a Triage Surface

External PRs are not a triage surface for these issue-tracker skills. Handle PRs through GitHub PR workflows.
