# Releasing a New Version

Use this flow to clean old release artifacts, build fresh binaries (linux/amd64 and darwin/arm64), generate checksums, and publish a GitHub release with assets.

```bash
# 1) set the new version in VERSION
echo "0.0.5" > VERSION

# 2) build release artifacts and checksums
make release

# 3) verify artifacts
ls -lh dist/
cat dist/checksums.txt

# 4) commit and push the version bump (if needed)
git add VERSION
git commit -m "Release v0.0.5"
git push

# 5) create and push tag
git tag -a v0.0.5 -m "v0.0.5"
git push origin v0.0.5

# 6) create GitHub release and upload artifacts
gh release create v0.0.5 \
  --title "v0.0.5" \
  --notes "Release v0.0.5" \
  dist/metrics-analyzer-0.0.5-linux-amd64 \
  dist/web-server-0.0.5-linux-amd64 \
  dist/metrics-analyzer-0.0.5-darwin-arm64 \
  dist/web-server-0.0.5-darwin-arm64 \
  dist/checksums.txt
```

