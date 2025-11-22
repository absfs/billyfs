# How to Create GitHub Issues from GITHUB_ISSUES.md

This guide walks through the process of creating GitHub issues from the prepared issue templates.

---

## Prerequisites

1. **Access to Repository**: You need write access to `github.com/absfs/billyfs`
2. **GitHub Account**: Logged in to GitHub
3. **Issue File**: `GITHUB_ISSUES.md` is available in the repository

---

## Method 1: Manual Creation via GitHub Web Interface (Recommended)

### Step 1: Navigate to Issues Page

```
https://github.com/absfs/billyfs/issues
```

Or from the repository:
1. Go to `https://github.com/absfs/billyfs`
2. Click the **"Issues"** tab at the top
3. Click the green **"New issue"** button

### Step 2: Copy Issue Content

Open `GITHUB_ISSUES.md` in your local editor or on GitHub:

```bash
# View locally
cat GITHUB_ISSUES.md

# Or open in your editor
code GITHUB_ISSUES.md  # VS Code
vim GITHUB_ISSUES.md   # Vim
nano GITHUB_ISSUES.md  # Nano
```

For each issue, copy the content between the markdown code blocks.

### Step 3: Fill in Issue Fields

#### For Issue #1 (Example):

**Title Field:**
```
Enable input validation in NewFS constructor
```

**Body Field:**
Copy everything from the Body section in GITHUB_ISSUES.md:

```markdown
## Problem
The `NewFS` constructor has essential validation code commented out (lines 25-38 in `billyfs.go`), allowing invalid inputs that can cause runtime errors or security issues.

## Current Code
[... rest of the body ...]
```

**Labels:**
Click "Labels" on the right sidebar and select:
- `bug`
- `critical`
- `security`

**Assignees (Optional):**
- Leave blank initially, or assign to yourself/maintainer

**Projects (Optional):**
- Can add to a project board if one exists

**Milestone (Optional):**
- Could create "v1.0 Production Ready" milestone

### Step 4: Submit the Issue

Click the green **"Submit new issue"** button at the bottom.

### Step 5: Repeat for Each Issue

Repeat steps 2-4 for all 17 issues.

---

## Method 2: Using GitHub CLI (Faster for Multiple Issues)

### Prerequisites

Install GitHub CLI:
```bash
# macOS
brew install gh

# Linux
sudo apt install gh  # Debian/Ubuntu
sudo dnf install gh  # Fedora

# Windows
winget install --id GitHub.cli
```

Authenticate:
```bash
gh auth login
```

### Create Issues from Command Line

Create a script to automate issue creation:

```bash
#!/bin/bash
# create-issues.sh

REPO="absfs/billyfs"

# Issue #1
gh issue create \
  --repo "$REPO" \
  --title "Enable input validation in NewFS constructor" \
  --body-file <(cat <<'EOF'
## Problem
The `NewFS` constructor has essential validation code commented out (lines 25-38 in `billyfs.go`), allowing invalid inputs that can cause runtime errors or security issues.

## Current Code
```go
func NewFS(fs absfs.SymlinkFileSystem, dir string) (*Filesystem, error) {
    //if dir == "" {
    //    return nil, os.ErrInvalid
    //}
    // ... more commented validation
```

## Expected Behavior
- Should validate that `dir` is not empty
- Should validate that `dir` is an absolute path
- Should validate that `dir` exists in the filesystem
- Should validate that `dir` is actually a directory (not a file)

## Acceptance Criteria
- [ ] Uncomment validation code
- [ ] Ensure all validation checks work correctly
- [ ] Add tests for each validation failure case
- [ ] Document validation requirements in godoc

## Related
Part of quality review findings - see PROJECT_QUALITY_REVIEW.md Issue #1
EOF
) \
  --label "bug,critical,security"

# Issue #2
gh issue create \
  --repo "$REPO" \
  --title "TempFile ignores dir parameter, always uses TempDir()" \
  --body-file <(cat <<'EOF'
[... copy body from GITHUB_ISSUES.md ...]
EOF
) \
  --label "bug,critical"

# Continue for all issues...
```

Make it executable and run:
```bash
chmod +x create-issues.sh
./create-issues.sh
```

### Simpler Alternative: Create Issues Interactively

```bash
# For each issue
gh issue create --repo absfs/billyfs

# This opens an interactive prompt:
# - Title: [paste title]
# - Body: [opens editor - paste body]
# - Save and submit
```

---

## Method 3: Using GitHub API (Most Automated)

### Using curl

```bash
#!/bin/bash

GITHUB_TOKEN="your_github_token_here"
REPO_OWNER="absfs"
REPO_NAME="billyfs"

# Issue #1
curl -X POST \
  -H "Authorization: token $GITHUB_TOKEN" \
  -H "Accept: application/vnd.github.v3+json" \
  https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/issues \
  -d '{
    "title": "Enable input validation in NewFS constructor",
    "body": "## Problem\nThe `NewFS` constructor has...",
    "labels": ["bug", "critical", "security"]
  }'
```

### Using Python Script

```python
#!/usr/bin/env python3
import requests
import os

GITHUB_TOKEN = os.environ.get('GITHUB_TOKEN')
REPO = 'absfs/billyfs'

issues = [
    {
        'title': 'Enable input validation in NewFS constructor',
        'body': '''## Problem
The `NewFS` constructor has essential validation code commented out...''',
        'labels': ['bug', 'critical', 'security']
    },
    # Add more issues...
]

headers = {
    'Authorization': f'token {GITHUB_TOKEN}',
    'Accept': 'application/vnd.github.v3+json'
}

for issue in issues:
    response = requests.post(
        f'https://api.github.com/repos/{REPO}/issues',
        headers=headers,
        json=issue
    )
    print(f"Created: {issue['title']} - {response.status_code}")
```

---

## Recommended Workflow

### Phase 1: Create Critical Issues First

Create issues in priority order:

1. **Day 1**: Create issues #1-3 (Critical bugs)
   ```bash
   gh issue create --title "Enable input validation in NewFS constructor" --label bug,critical,security
   gh issue create --title "TempFile ignores dir parameter, always uses TempDir()" --label bug,critical
   gh issue create --title "Replace deprecated rand.Seed in TempFile implementation" --label bug,medium,technical-debt
   ```

2. **Day 1**: Create issue #6 (CI/CD)
   ```bash
   gh issue create --title "Add GitHub Actions workflow for automated testing" --label infrastructure,ci-cd,high-priority
   ```

3. **Day 2**: Create issues #4-5 (Testing)
   ```bash
   gh issue create --title "Add comprehensive test coverage for filesystem operations" --label testing,critical,good-first-issue
   gh issue create --title "Add runnable example tests" --label testing,documentation,good-first-issue
   ```

4. **Day 3+**: Create remaining issues as needed

### Phase 2: Organize Issues

After creating issues:

1. **Create Labels** (if they don't exist):
   - Go to Issues → Labels
   - Create: `bug`, `critical`, `high-priority`, `medium-priority`, `testing`, `documentation`, `infrastructure`, `good-first-issue`, etc.

2. **Create Milestone**:
   - Go to Issues → Milestones → New Milestone
   - Name: "v1.0 - Production Ready"
   - Due date: 2-3 weeks out
   - Assign critical issues to this milestone

3. **Create Project Board** (Optional):
   - Go to Projects → New Project
   - Template: "Basic Kanban" or "Automated Kanban"
   - Columns: "To Do", "In Progress", "Done"
   - Add issues to board

### Phase 3: Prioritize and Track

1. **Pin Critical Issues**:
   - Open issue #1, #2, #3
   - Click "..." menu → "Pin issue"
   - These appear at top of issues list

2. **Add Issue Numbers to Cross-References**:
   - After creating all issues, edit to add cross-references
   - Example: "See also #4, #5" in issue #6

3. **Create Meta-Issue** (Optional):
   ```markdown
   Title: Production Quality Roadmap

   Body:
   ## Critical Bugs
   - [ ] #1 Enable input validation
   - [ ] #2 Fix TempFile bug
   - [ ] #3 Replace deprecated rand.Seed

   ## Testing
   - [ ] #4 Comprehensive test suite
   - [ ] #5 Example tests

   [... etc ...]
   ```

---

## Tips and Best Practices

### 1. Don't Create All Issues at Once

**Why**: Can overwhelm maintainers and contributors

**Better Approach**:
- Start with 5-7 critical issues
- Add more as work progresses
- Prevents "issue fatigue"

### 2. Customize Issue Templates

Create `.github/ISSUE_TEMPLATE/bug_report.md`:
```markdown
---
name: Bug Report
about: Report a bug
labels: bug
---

## Problem
[Description]

## Expected Behavior
[What should happen]

## Acceptance Criteria
- [ ] [Criteria 1]
```

### 3. Use Issue Templates in GitHub

Go to Settings → Features → Issues → Set up templates

### 4. Tag for Newcomers

Use `good-first-issue` label for:
- Issue #5 (Example tests)
- Issue #10 (Remove unused dependency)
- Issue #13 (CHANGELOG.md)

### 5. Break Large Issues into Smaller Ones

Issue #4 (Comprehensive tests) is very large. Consider splitting:
- Issue #4a: Add basic file operation tests
- Issue #4b: Add directory operation tests
- Issue #4c: Add symlink operation tests
- etc.

### 6. Link Issues to PRs

When creating pull requests:
```
git commit -m "Fix TempFile to honor dir parameter

Fixes #2"
```

The `Fixes #2` automatically closes the issue when PR merges.

---

## Quick Reference: Issue Creation Commands

### GitHub CLI One-Liners

```bash
# Critical bugs
gh issue create -R absfs/billyfs -t "Enable input validation in NewFS constructor" -l bug,critical,security
gh issue create -R absfs/billyfs -t "TempFile ignores dir parameter" -l bug,critical
gh issue create -R absfs/billyfs -t "Replace deprecated rand.Seed" -l bug,medium

# Infrastructure
gh issue create -R absfs/billyfs -t "Add GitHub Actions CI/CD pipeline" -l infrastructure,high-priority

# Testing
gh issue create -R absfs/billyfs -t "Add comprehensive test suite" -l testing,critical,good-first-issue
gh issue create -R absfs/billyfs -t "Add example tests" -l testing,documentation,good-first-issue
```

### View Created Issues

```bash
# List all issues
gh issue list -R absfs/billyfs

# View specific issue
gh issue view 1 -R absfs/billyfs

# List issues by label
gh issue list -R absfs/billyfs -l bug
gh issue list -R absfs/billyfs -l critical
```

---

## Automated Bulk Creation Script

Create `bulk-create-issues.sh`:

```bash
#!/bin/bash
set -e

REPO="absfs/billyfs"

# Check gh CLI is installed
if ! command -v gh &> /dev/null; then
    echo "GitHub CLI (gh) is not installed. Install it first."
    exit 1
fi

# Check authentication
if ! gh auth status &> /dev/null; then
    echo "Not authenticated. Run: gh auth login"
    exit 1
fi

echo "Creating issues for $REPO..."

# Issue #1
echo "Creating issue #1..."
gh issue create -R "$REPO" \
  --title "Enable input validation in NewFS constructor" \
  --label "bug,critical,security" \
  --body "$(cat <<'EOF'
## Problem
The \`NewFS\` constructor has essential validation code commented out (lines 25-38 in \`billyfs.go\`), allowing invalid inputs that can cause runtime errors or security issues.

## Current Code
\`\`\`go
func NewFS(fs absfs.SymlinkFileSystem, dir string) (*Filesystem, error) {
    //if dir == "" {
    //    return nil, os.ErrInvalid
    //}
    // ... more commented validation
\`\`\`

## Expected Behavior
- Should validate that \`dir\` is not empty
- Should validate that \`dir\` is an absolute path
- Should validate that \`dir\` exists in the filesystem
- Should validate that \`dir\` is actually a directory (not a file)

## Acceptance Criteria
- [ ] Uncomment validation code
- [ ] Ensure all validation checks work correctly
- [ ] Add tests for each validation failure case
- [ ] Document validation requirements in godoc

## Related
Part of quality review findings - see PROJECT_QUALITY_REVIEW.md Issue #1
EOF
)"

# Issue #2
echo "Creating issue #2..."
gh issue create -R "$REPO" \
  --title "TempFile ignores dir parameter, always uses TempDir()" \
  --label "bug,critical" \
  --body "$(cat <<'EOF'
## Problem
The \`TempFile\` method completely ignores its \`dir\` parameter and always creates files in the system temp directory. This violates the billy.Filesystem interface contract.

## Current Code
\`\`\`go
func (f *Filesystem) TempFile(dir string, prefix string) (billy.File, error) {
    rand.Seed(time.Now().UnixNano())
    p := path.Join(f.fs.TempDir(), prefix+"_"+randSeq(5))  // ❌ Ignores 'dir'
    // ...
}
\`\`\`

## Expected Behavior
Per billy.Filesystem specification:
- If \`dir\` is non-empty, create temp file in that directory
- If \`dir\` is empty, use default temp directory
- File should be created with unique name

## Acceptance Criteria
- [ ] Honor \`dir\` parameter when provided
- [ ] Fall back to TempDir() only when \`dir\` is empty
- [ ] Add test cases for both scenarios
- [ ] Verify compatibility with go-git usage patterns

## Related
Part of quality review findings - see PROJECT_QUALITY_REVIEW.md Issue #2
EOF
)"

# Add more issues as needed...

echo "✓ Issues created successfully!"
echo "View at: https://github.com/$REPO/issues"
```

Make executable and run:
```bash
chmod +x bulk-create-issues.sh
./bulk-create-issues.sh
```

---

## Summary

**Easiest Method**: GitHub Web Interface (manual copy/paste)
**Fastest Method**: GitHub CLI with prepared script
**Most Flexible**: GitHub CLI interactive mode

**Recommended**: Start with 5-7 critical issues manually, then automate the rest if needed.
