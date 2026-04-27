# Lunie Docs Site

This app hosts the Lunie documentation website using Docusaurus.

The content source is the repository-level `docs/` directory.

## Local Development

From repository root:

```bash
pnpm docs:dev
```

Or directly:

```bash
pnpm -C apps/docs start
```

## Build

```bash
pnpm docs:build
```

## Serve Built Site

```bash
pnpm docs:serve
```

## Deployment

Deployment is handled by `.github/workflows/docs-deploy.yml`.
