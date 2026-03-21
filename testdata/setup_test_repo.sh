#!/bin/bash
# Creates a test git repository with various types of evil merges.
# Usage: ./setup_test_repo.sh <target_dir>

set -e

TARGET="${1:?Usage: $0 <target_dir>}"
rm -rf "$TARGET"
mkdir -p "$TARGET"
cd "$TARGET"

git init
git config user.email "test@example.com"
git config user.name "Test User"

# Initial commit
echo "hello world" > file1.txt
echo "base content" > file2.txt
echo "shared code" > shared.txt
git add -A
git commit -m "Initial commit"

# === Evil Merge Type 1: Change in file untouched by both branches ===
git checkout -b feature-a
echo "feature A changes" > file1.txt
git add file1.txt
git commit -m "Feature A: modify file1"

git checkout main
git checkout -b feature-b
echo "feature B changes" > file2.txt
git add file2.txt
git commit -m "Feature B: modify file2"

git checkout main
git merge feature-a --no-ff -m "Merge feature-a"
# Now merge feature-b but sneak in an evil change to shared.txt
git merge feature-b --no-ff -m "Merge feature-b (evil: modifies shared.txt)"
# Amend the merge to include an evil change
echo "evil change - not in any branch" > shared.txt
git add shared.txt
git commit --amend -m "Merge feature-b (evil: modifies shared.txt)" --no-edit 2>/dev/null || \
git commit --amend -m "Merge feature-b (evil: modifies shared.txt)"

# === Evil Merge Type 2: New file added only in merge ===
git checkout -b feature-c
echo "feature C" > featureC.txt
git add featureC.txt
git commit -m "Feature C: add featureC.txt"

git checkout main
git checkout -b feature-d
echo "feature D" > featureD.txt
git add featureD.txt
git commit -m "Feature D: add featureD.txt"

git checkout main
git merge feature-c --no-ff -m "Merge feature-c"
git merge feature-d --no-ff -m "Merge feature-d (evil: adds secret.env)"
# Add evil new file in merge
echo "SECRET_KEY=evil123" > secret.env
git add secret.env
git commit --amend -m "Merge feature-d (evil: adds secret.env)"

# === Evil Merge Type 3: Clean merge (no evil changes) — control case ===
git checkout -b feature-clean
echo "clean feature" > clean.txt
git add clean.txt
git commit -m "Clean feature"

git checkout main
git merge feature-clean --no-ff -m "Merge feature-clean (no evil changes)"

# === Evil Merge Type 4: Conflict zone with extra changes ===
git checkout -b branch-left
echo "left version" > conflict.txt
git add conflict.txt
git commit -m "Left: add conflict.txt"

git checkout main
git checkout -b branch-right
echo "right version" > conflict.txt
git add conflict.txt
git commit -m "Right: add conflict.txt"

git checkout main
git merge branch-left --no-ff -m "Merge branch-left"
# Merge branch-right will conflict
git merge branch-right --no-ff -m "Merge branch-right (conflict resolution)" || true
echo "resolved version plus extra evil line" > conflict.txt
git add conflict.txt
git commit -m "Merge branch-right (conflict resolution with extra changes)"

echo ""
echo "Test repository created at: $TARGET"
echo "Expected evil merges: 3 (feature-b merge, feature-d merge, branch-right merge)"
echo "Expected clean merges: 3 (feature-a, feature-c, feature-clean)"
