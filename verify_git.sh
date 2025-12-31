#!/bin/bash
set -e

# Cleanup
rm -rf /tmp/monarch-git-verify

PROJECT_ROOT=$(pwd)

# Prepare directories
mkdir -p /tmp/monarch-git-verify

# Setup 'behind' repo manually
echo "Setting up 'behind' repo..."
git clone https://github.com/octocat/Hello-World.git /tmp/monarch-git-verify/behind
cd /tmp/monarch-git-verify/behind
# Reset to previous commit so 'update: true' has something to do
git reset --hard HEAD~1
echo "Repo reset to HEAD~1. Current HEAD:"
git rev-parse HEAD
OLD_HEAD=$(git rev-parse HEAD)

cd -

# Run Monarch
echo "Running Monarch Apply..."
./monarch apply monarch-git-verify.yaml

# Verification
echo "Verifying results..."

# Check update
cd /tmp/monarch-git-verify/behind
NEW_HEAD=$(git rev-parse HEAD)
echo "Updated HEAD: $NEW_HEAD"
if [ "$NEW_HEAD" == "$OLD_HEAD" ]; then
  echo "❌ Update FAILED: Head is still at old commit ($OLD_HEAD)."
  exit 1
else
  echo "✅ Update SUCCESS: Head moved forward to $NEW_HEAD."
fi

# Check commit checkout
cd /tmp/monarch-git-verify/commit
COMMIT_HEAD=$(git rev-parse HEAD)
echo "Commit specific HEAD: $COMMIT_HEAD"
if [ "$COMMIT_HEAD" == "7fd1a60b01f91b314f59955a4e4d4e80d8edf11d" ]; then
    echo "✅ Commit Checkout SUCCESS."
else
    echo "❌ Commit Checkout FAILED."
    exit 1
fi

# 4. Check Remote Mismatch (Safety)
echo "Setting up 'mismatch' repo..."
git clone https://github.com/octocat/Hello-World.git /tmp/monarch-git-verify/mismatch

echo "Running Monarch Apply for Mismatch (Should Fail)..."
set +e
$PROJECT_ROOT/monarch apply $PROJECT_ROOT/monarch-git-mismatch.yaml
EXIT_CODE=$?
set -e

if [ $EXIT_CODE -ne 0 ]; then
    echo "✅ Remote Mismatch Safety Check SUCCESS: Monarch failed as expected."
else
    echo "❌ Remote Mismatch Safety Check FAILED: Monarch should have errored but succeeded."
    exit 1
fi

echo "All Git Tests Passed!"
