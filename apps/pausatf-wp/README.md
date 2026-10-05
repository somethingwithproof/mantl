# PAUSATF WordPress (Bedrock-style)

This app is a Composer-managed WordPress project for the pausatf.org site. It uses a Bedrock-style layout with environment variables, Composer-managed core/plugins, and CI.

## Layout
- `web/wp/` – WordPress core (installed by Composer)
- `web/app/` – wp-content (themes, plugins, mu-plugins, uploads)
- `config/` – shared config and per-environment overrides
- `.env` – environment variables (copy from `.env.example`)

## Quick start (local)
1. Install the PHP and Composer versions pinned in this directory:
   ```sh
   mise install vfox:jdx/vfox-php github:composer/composer
   ```
   The PHP backend builds from source; install its [native build prerequisites](https://github.com/jdx/vfox-php#requirements).
2. Install deps:
   ```sh
   mise exec -- php "$(mise where github:composer/composer@2.10.3)/composer.phar" install
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
- Commit `composer.lock` when updating dependencies. Regenerate it with the
  pinned Composer runtime using `update --no-install --no-scripts --no-plugins`;
  validate it with `validate --strict --no-check-publish` before review.

WordPress 6.8 or newer owns bcrypt password hashing. The abandoned
`roots/wp-password-bcrypt` plugin is intentionally absent; existing passwords
remain compatible. See [Roots’ migration guidance](https://roots.io/sunsetting-wp-password-bcrypt-with-wordpress-6-8/).
