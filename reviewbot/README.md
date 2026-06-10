# ReviewBot — AI Code Review as a Service 🦾

DeepSeek-powered GitHub PR review bot. Auto-reviews every PR, posts inline comments.

## How it works

```
GitHub PR → Webhook → ReviewBot → DeepSeek → PR Comments
```

1. You push a PR
2. GitHub sends webhook to ReviewBot
3. ReviewBot fetches the diff, filters out noise (lock files, binaries)
4. Sends to DeepSeek for review
5. Posts inline comments back to the PR

## Quick Start

```bash
# 1. Set up env
export DEEPSEEK_API_KEY=sk-...
export GITHUB_TOKEN=ghp_...
export WEBHOOK_SECRET=my-random-secret

# 2. Install & run
uv sync
uv run python main.py --port 8080

# 3. Expose to internet (for GitHub webhook)
# Option A: ngrok
ngrok http 8080

# Option B: Deploy to Railway/Render/Fly.io

# 4. Configure GitHub webhook
#    Repo → Settings → Webhooks → Add webhook
#    URL: https://your-domain/webhook
#    Secret: my-random-secret
#    Events: Pull requests
```

## Pricing (SaaS model)

| Tier | Price | What |
|------|-------|------|
| Free | $0 | Public repos, 50 reviews/month |
| Pro | $9/mo | Private repos, unlimited reviews |
| Team | $29/mo | 5 repos, priority queue, custom rules |

Your cost per review: ~$0.002 (DeepSeek, ~2000 tokens avg)

## Tech Stack

- **Python 3.11+** / Flask
- **DeepSeek API** (1/50th the cost of GPT-4)
- **GitHub REST API** (webhooks + reviews)

## Files

```
reviewbot/
├── main.py          # Flask server + review logic
├── pyproject.toml   # Dependencies (flask, requests)
└── README.md        # This file
```

## Next Steps

- [ ] Test with a real GitHub repo
- [ ] Add review approval/rejection tracking
- [ ] Add language-specific review rules (Go, Java, Python)
- [ ] Build landing page
- [ ] Set up Stripe billing
- [ ] Publish to GitHub Marketplace
