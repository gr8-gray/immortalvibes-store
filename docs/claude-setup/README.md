# Claude Code starter

This sets up Claude Code in your terminal the way the previous maintainer of Immortal Vibes runs it:
a global instruction file every session reads, plus a personal **mistakes ledger** that grows each
time you or Claude get something wrong, so the same mistake doesn't happen twice.

It takes about 15 minutes. Run the commands from the root of your clone of this repo
(`git clone https://github.com/gr8-gray/immortalvibes-store.git && cd immortalvibes-store`).

## What's in this folder

| File | What it is |
|---|---|
| `README.md` | This guide |
| `global-CLAUDE.md.template` | Your global instructions. Every Claude Code session on your machine reads it. |
| `mistakes-ledger.template.md` | Your ledger. It starts empty, and you and Claude add to it. |

(These files are also handed out as a zip with the templates already named `CLAUDE.md` and
`mistakes-ledger.md`.)

## 1. Account

Claude Code needs your own paid Claude plan (Pro or Max) or Anthropic API credits. Use your own
account, not a shared one.

## 2. Install

**macOS / Linux / WSL:**
```bash
curl -fsSL https://claude.ai/install.sh | bash
```

**Windows (PowerShell):** install [Git for Windows](https://git-scm.com/downloads/win) first. Claude Code on
native Windows uses its Git Bash. Then:
```powershell
irm https://claude.ai/install.ps1 | iex
```

**Or with Node 18+ installed:** `npm install -g @anthropic-ai/claude-code`

If the install commands have changed, follow the official setup page:
https://docs.claude.com/en/docs/claude-code/setup

Check it worked, then sign in:
```bash
claude --version
claude          # opens a session and walks you through login
```

## 3. Install your global instructions

Copy `global-CLAUDE.md.template` into your Claude folder as `CLAUDE.md`:

| OS | Destination |
|---|---|
| macOS / Linux / WSL | `~/.claude/CLAUDE.md` |
| Windows | `%USERPROFILE%\.claude\CLAUDE.md` |

```bash
# macOS / Linux / WSL
mkdir -p ~/.claude && cp docs/claude-setup/global-CLAUDE.md.template ~/.claude/CLAUDE.md
```

```powershell
# Windows (PowerShell)
New-Item -ItemType Directory -Force "$HOME\.claude" | Out-Null
Copy-Item docs\claude-setup\global-CLAUDE.md.template "$HOME\.claude\CLAUDE.md"
```

Open it and fill in the four `<<...>>` placeholders at the top: your name, your background, what
you're working on, and where your ledger lives.

## 4. Set up your mistakes ledger

Keep it in a **private** git repo of your own, so it survives a new laptop and has history:

```bash
# macOS / Linux / WSL
mkdir -p ~/dev/claude-memory && cp docs/claude-setup/mistakes-ledger.template.md ~/dev/claude-memory/mistakes-ledger.md
cd ~/dev/claude-memory && git init && git add . && git commit -m "ledger: start"
```

```powershell
# Windows (PowerShell)
New-Item -ItemType Directory -Force "$HOME\dev\claude-memory" | Out-Null
Copy-Item docs\claude-setup\mistakes-ledger.template.md "$HOME\dev\claude-memory\mistakes-ledger.md"
Set-Location "$HOME\dev\claude-memory"; git init; git add .; git commit -m "ledger: start"
```

Then create a **private** repo on GitHub and push it.

Put that path in the `<<LEDGER PATH>>` placeholder in your `CLAUDE.md`.

**Never make the ledger repo public.** Entries often name customers, accounts or secrets by accident.

## 5. First session on Immortal Vibes

From the root of your clone:

```bash
claude
```

The repo's `CLAUDE.md` loads automatically in every session you start inside it, and it imports
`docs/lessons/RULES.md` and `docs/lessons/OPEN-WORK.md`. The incident write-ups in `docs/lessons/incidents/`
are read on demand.

A good first prompt: *"Read CLAUDE.md and docs/lessons/RULES.md, then explain how a change gets to
production and how I know it worked."*

## Habits that make this work

- **When Claude gets something wrong and you correct it, say "add that to the ledger."** That one
  habit is the whole system.
- **Plan before big changes:** press `Shift+Tab` to switch to plan mode, then read the plan before it edits.
- **Run a shell command in the session** with `!`, for example `! git status`.
- **Starting a different task?** Use `/clear`. Stale context causes confident wrong answers.
- **"Done" means you saw the proof:** test output, a green CI run, the live site. Not Claude saying it's done.
- **Merged is not deployed.** Check the deploy actually ran and passed.
