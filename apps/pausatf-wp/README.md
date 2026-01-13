# PAUSATF WordPress (Bedrock-style)

This app is a Composer-managed WordPress project for the pausatf.org site. It uses a Bedrock-style layout with environment variables, Composer-managed core/plugins, and CI.

## Layout
- `web/wp/` – WordPress core (installed by Composer)
- `web/app/` – wp-content (themes, plugins, mu-plugins, uploads)
- `config/` – shared config and per-environment overrides
- `.env` – environment variables (copy from `.env.example`)

## Quick start (local)
1. Prereqs: PHP 8.2/8.3, Composer, Node 20+.
2. Install deps:
   ```sh
   composer install
   ```
3. Configure environment:
   - Copy `.env.example` to `.env` and set values.
4. Set document root to `web/` when serving locally.

## Deploy (high level)
- Build on CI and rsync/symlink a release to the server.
- Keep `web/app/uploads` outside releases at `/var/www/pausatf/shared/uploads`.

## CI
- Runs PHPCS (WordPressCS) and PHPStan. JS/CSS linting via ESLint/Stylelint if present.

## Notes
- Do not commit real secrets. `.env` is ignored. Use `.env.example` as a template.
