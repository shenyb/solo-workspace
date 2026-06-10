#!/usr/bin/env python3
"""
ReviewBot — AI-powered GitHub code review bot (DeepSeek edition).

Receives GitHub webhooks, reviews PR diffs with DeepSeek, and posts
inline review comments.

Quick start:
  export DEEPSEEK_API_KEY=sk-...
  export GITHUB_TOKEN=ghp_...
  export WEBHOOK_SECRET=your-secret

  uv run python main.py --port 8080

  # Then configure GitHub webhook:
  #   URL: https://your-domain.com/webhook
  #   Secret: your-secret
  #   Events: Pull requests

Architecture:
  GitHub PR event → Flask webhook → DeepSeek review → PR comments
"""

import hashlib
import hmac
import json
import logging
import os
import re
import sys
import time
from argparse import ArgumentParser
from typing import Optional

import requests
from flask import Flask, Request, jsonify, request

# ── Config ─────────────────────────────────────────────────────────────────

DEEPSEEK_API_KEY = os.environ.get("DEEPSEEK_API_KEY", "")
GITHUB_TOKEN = os.environ.get("GITHUB_TOKEN", "")
WEBHOOK_SECRET = os.environ.get("WEBHOOK_SECRET", "")

DEEPSEEK_BASE = "https://api.deepseek.com"
DEEPSEEK_MODEL = os.environ.get("REVIEWBOT_MODEL", "deepseek-chat")

# Files to skip — binary, generated, config
SKIP_EXTENSIONS = {
    ".lock", ".sum", ".svg", ".png", ".jpg", ".jpeg", ".gif", ".ico",
    ".woff", ".woff2", ".ttf", ".eot", ".mp4", ".mp3", ".webp",
    ".pb.go",  # protobuf generated
    ".gen.go",  # go generate output
}
SKIP_PATTERNS = [
    r"^package-lock\.json$",
    r"^yarn\.lock$",
    r"^pnpm-lock\.yaml$",
    r"^go\.sum$",
    r"^Cargo\.lock$",
    r"^Gemfile\.lock$",
    r"^poetry\.lock$",
    r"^Pipfile\.lock$",
    r"\.min\.(js|css)$",
]

# Maximum diff size to send (characters)
MAX_DIFF_SIZE = 8000

# ── Logging ─────────────────────────────────────────────────────────────────

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    stream=sys.stderr,
)
log = logging.getLogger("reviewbot")

# ── Flask app ───────────────────────────────────────────────────────────────

app = Flask(__name__)


# ── GitHub helpers ──────────────────────────────────────────────────────────

def verify_signature(req: Request) -> bool:
    """Verify the GitHub webhook signature."""
    if not WEBHOOK_SECRET:
        log.warning("WEBHOOK_SECRET not set — skipping signature verification")
        return True

    sig = req.headers.get("X-Hub-Signature-256", "")
    if not sig:
        log.warning("No signature header found")
        return False

    mac = hmac.new(
        WEBHOOK_SECRET.encode(),
        req.get_data(),
        hashlib.sha256,
    )
    expected = f"sha256={mac.hexdigest()}"
    return hmac.compare_digest(sig, expected)


def github_api(path: str, method: str = "GET", body: dict = None) -> dict:
    """Call the GitHub REST API."""
    url = f"https://api.github.com{path}"
    headers = {
        "Authorization": f"Bearer {GITHUB_TOKEN}",
        "Accept": "application/vnd.github+json",
        "X-GitHub-Api-Version": "2022-11-28",
    }
    if method == "GET":
        resp = requests.get(url, headers=headers, timeout=30)
    elif method == "POST":
        headers["Content-Type"] = "application/json"
        resp = requests.post(url, headers=headers, json=body, timeout=30)
    else:
        raise ValueError(f"Unsupported method: {method}")

    resp.raise_for_status()
    return resp.json()


def get_pr_diff(owner: str, repo: str, pr_number: int) -> str:
    """Get the raw diff for a PR."""
    url = f"https://api.github.com/repos/{owner}/{repo}/pulls/{pr_number}"
    headers = {
        "Authorization": f"Bearer {GITHUB_TOKEN}",
        "Accept": "application/vnd.github.v3.diff",
    }
    resp = requests.get(url, headers=headers, timeout=30)
    resp.raise_for_status()
    return resp.text


def should_skip_file(filename: str) -> bool:
    """Check if a file should be skipped for review."""
    ext = os.path.splitext(filename)[1].lower()
    if ext in SKIP_EXTENSIONS:
        return True
    basename = os.path.basename(filename)
    for pattern in SKIP_PATTERNS:
        if re.match(pattern, basename):
            return True
    return False


def filter_diff(diff: str) -> str:
    """Remove files we don't want to review from the diff."""
    lines = diff.split("\n")
    filtered = []
    skip_current = False

    for line in lines:
        # Detect file headers: "diff --git a/file b/file"
        if line.startswith("diff --git "):
            parts = line.split(" ")
            if len(parts) >= 4:
                filename = parts[3][2:]  # strip "b/" prefix
                skip_current = should_skip_file(filename)
        if not skip_current:
            filtered.append(line)

    result = "\n".join(filtered)
    if len(result) > MAX_DIFF_SIZE:
        result = result[:MAX_DIFF_SIZE] + "\n...[diff truncated]"
    return result


def post_pr_review(
    owner: str,
    repo: str,
    pr_number: int,
    commit_id: str,
    comments: list[dict],
    summary: str = "",
):
    """Post a review with inline comments to a PR."""
    body = {
        "commit_id": commit_id,
        "event": "COMMENT",
        "comments": comments,
    }
    if summary:
        body["body"] = summary
    return github_api(
        f"/repos/{owner}/{repo}/pulls/{pr_number}/reviews",
        method="POST",
        body=body,
    )


# ── DeepSeek review logic ─────────────────────────────────────────────────

REVIEW_SYSTEM_PROMPT = """You are an expert code reviewer. Review the provided git diff and give constructive feedback.

Guidelines:
1. Focus on bugs, security issues, and logic errors — these are your top priority.
2. Flag code smells: overly complex functions, duplicated code, missing error handling.
3. Check for performance issues: N+1 queries, unnecessary allocations, blocking operations.
4. Note style/convention issues only if they're genuinely confusing.
5. Be constructive and specific — suggest concrete fixes, not vague complaints.
6. If the code looks fine, say so — no need to nitpick.

Output format — return a JSON object:
{
  "summary": "Brief overall assessment (1-3 sentences)",
  "comments": [
    {
      "file": "path/to/file.go",
      "line": 42,
      "severity": "high|medium|low|info",
      "body": "Specific feedback with suggested fix"
    }
  ]
}

Rules:
- Output ONLY the JSON object — no markdown, no explanation.
- line number should be the NEW line number (right side of diff, after the +).
- Only comment on real issues — skip files with only formatting changes.
- If the diff is empty or trivially small, return {"summary": "...", "comments": []}.
- Maximum 10 comments per review."""


def review_diff(diff: str, title: str = "") -> dict:
    """Send diff to DeepSeek for review. Returns parsed review result."""
    if not DEEPSEEK_API_KEY:
        return {"summary": "DEEPSEEK_API_KEY not configured", "comments": []}

    # Fallback for empty diffs
    if not diff.strip() or len(diff.strip()) < 10:
        return {"summary": "Diff is empty or too small to review.", "comments": []}

    user_prompt = f"PR title: {title}\n\nDiff to review:\n```diff\n{diff}\n```"

    body = {
        "model": DEEPSEEK_MODEL,
        "messages": [
            {"role": "system", "content": REVIEW_SYSTEM_PROMPT},
            {"role": "user", "content": user_prompt},
        ],
        "temperature": 0.3,
        "max_tokens": 4096,
    }

    log.info(f"  Sending {len(diff)} chars to DeepSeek...")
    t0 = time.time()

    resp = requests.post(
        f"{DEEPSEEK_BASE}/v1/chat/completions",
        headers={
            "Authorization": f"Bearer {DEEPSEEK_API_KEY}",
            "Content-Type": "application/json",
        },
        json=body,
        timeout=120,
    )
    resp.raise_for_status()
    data = resp.json()
    elapsed = time.time() - t0

    content = data["choices"][0]["message"]["content"]
    tokens = data.get("usage", {}).get("total_tokens", 0)
    log.info(f"  Review complete in {elapsed:.1f}s ({tokens} tokens)")

    # Parse the JSON response
    try:
        # Strip markdown code fences if present
        content = re.sub(r"^```(?:json)?\s*\n?", "", content.strip())
        content = re.sub(r"\n```\s*$", "", content.strip())
        result = json.loads(content)
        return result
    except json.JSONDecodeError as e:
        log.error(f"Failed to parse review response: {e}")
        log.error(f"Raw response: {content[:500]}")
        return {
            "summary": f"Review completed but failed to parse AI response.",
            "comments": [],
            "_raw": content[:1000],
        }


# ── Webhook handler ────────────────────────────────────────────────────────

@app.route("/webhook", methods=["POST"])
def handle_webhook():
    """Handle GitHub webhook events."""
    # Verify signature
    if not verify_signature(request):
        return jsonify({"error": "Invalid signature"}), 403

    event = request.headers.get("X-GitHub-Event", "")
    payload = request.get_json()

    if event != "pull_request":
        log.info(f"Ignoring event: {event}")
        return jsonify({"status": "ignored", "event": event})

    action = payload.get("action", "")
    log.info(f"PR event: {action}")

    # Only review on open, reopen, or new commits
    if action not in ("opened", "reopened", "synchronize"):
        log.info(f"Skipping action: {action}")
        return jsonify({"status": "skipped", "action": action})

    pr = payload["pull_request"]
    repo = payload["repository"]
    owner = repo["owner"]["login"]
    repo_name = repo["name"]
    pr_number = pr["number"]
    pr_title = pr["title"]
    commit_id = pr["head"]["sha"]

    log.info(f"Reviewing PR: {owner}/{repo_name}#{pr_number} — {pr_title}")

    try:
        # Get the diff
        diff = get_pr_diff(owner, repo_name, pr_number)
        filtered = filter_diff(diff)
        log.info(f"  Diff: {len(diff)} chars → {len(filtered)} chars after filtering")

        # Review it
        review = review_diff(filtered, title=pr_title)

        # Post comments
        comments = review.get("comments", [])
        if comments:
            log.info(f"  Posting {len(comments)} review comments")
            post_pr_review(owner, repo_name, pr_number, commit_id, comments,
                          summary=review.get("summary", ""))
        else:
            log.info("  No issues found, skipping comment")

        return jsonify({
            "status": "reviewed",
            "pr": f"{owner}/{repo_name}#{pr_number}",
            "comments": len(comments),
            "summary": review.get("summary", ""),
        })

    except requests.HTTPError as e:
        log.error(f"GitHub API error: {e}")
        return jsonify({"error": str(e)}), 500
    except Exception as e:
        log.exception(f"Review failed: {e}")
        return jsonify({"error": str(e)}), 500


@app.route("/health", methods=["GET"])
def health():
    """Health check endpoint."""
    status = {
        "status": "ok",
        "deepseek": "configured" if DEEPSEEK_API_KEY else "missing",
        "github": "configured" if GITHUB_TOKEN else "missing",
        "webhook_secret": "configured" if WEBHOOK_SECRET else "missing",
    }
    return jsonify(status)


@app.route("/", methods=["GET"])
def index():
    return """<h1>ReviewBot 🦾</h1>
<p>AI-powered code review via DeepSeek.</p>
<p>Endpoint: <code>POST /webhook</code></p>
<p><a href="/health">Health check</a></p>"""


# ── CLI ───────────────────────────────────────────────────────────────────

def main():
    parser = ArgumentParser(description="ReviewBot — AI code review server")
    parser.add_argument("--port", "-p", type=int, default=8080,
                       help="Port to listen on (default: 8080)")
    parser.add_argument("--host", default="0.0.0.0",
                       help="Host to bind (default: 0.0.0.0)")
    parser.add_argument("--debug", action="store_true",
                       help="Enable Flask debug mode")
    args = parser.parse_args()

    # Validate config
    missing = []
    if not GITHUB_TOKEN:
        missing.append("GITHUB_TOKEN")
    if not DEEPSEEK_API_KEY:
        missing.append("DEEPSEEK_API_KEY")
    if missing:
        log.warning(f"Missing env vars: {', '.join(missing)}")
        log.warning("Set them before expecting full functionality.")

    log.info(f"Starting ReviewBot on {args.host}:{args.port}")
    app.run(host=args.host, port=args.port, debug=args.debug)


if __name__ == "__main__":
    main()
