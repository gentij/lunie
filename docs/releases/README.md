# Release Notes

This directory stores the manual release note template and curated release descriptions.

Suggested workflow:

1. Tag the release so CI builds and uploads assets.
2. Wait for the GitHub release to be created by GoReleaser.
3. Copy the matching `docs/releases/vX.Y.Z.md` contents into the GitHub release body.
4. Edit the GitHub release notes once the packaging workflows have succeeded.

Files:

- `RELEASE_TEMPLATE.md` - reusable structure for future releases
- `v1.0.0.md` - curated notes for the first public Lunie release
- `v1.1.0.md` - curated notes for the durable identifier and CLI/TUI UX update
