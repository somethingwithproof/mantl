<?php

use function Env\env;

// Load environment variables
if (file_exists(dirname(__DIR__).'/.env')) {
    \Env\Env::init();
}

// Set up environment
define('WP_ENV', env('WP_ENV') ?: 'production');

// URLs
define('WP_HOME', env('WP_HOME'));
define('WP_SITEURL', env('WP_SITEURL'));

// Custom content directory
define('CONTENT_DIR', '/app');
define('WP_CONTENT_DIR', dirname(__DIR__).'/web'.CONTENT_DIR);
define('WP_CONTENT_URL', WP_HOME.CONTENT_DIR);

// DB settings
define('DB_NAME', env('DB_NAME'));
define('DB_USER', env('DB_USER'));
define('DB_PASSWORD', env('DB_PASSWORD'));
define('DB_HOST', env('DB_HOST') ?: 'localhost');
define('DB_CHARSET', env('DB_CHARSET') ?: 'utf8mb4');
define('DB_COLLATE', env('DB_COLLATE') ?: '');
$table_prefix = env('TABLE_PREFIX') ?: 'wp_';

// Security keys & salts
foreach ([
    'AUTH_KEY', 'SECURE_AUTH_KEY', 'LOGGED_IN_KEY', 'NONCE_KEY',
    'AUTH_SALT', 'SECURE_AUTH_SALT', 'LOGGED_IN_SALT', 'NONCE_SALT',
] as $key) {
    $value = env($key);
    if ($value) {
        define($key, $value);
    } elseif (WP_ENV === 'production') {
        throw new \RuntimeException('Required security key not set: ' . $key);
    }
}

// Security hardening
define('DISALLOW_FILE_EDIT', env('DISALLOW_FILE_EDIT') ?: true);
define('DISALLOW_FILE_MODS', env('DISALLOW_FILE_MODS') ?: false);
define('FORCE_SSL_ADMIN', env('FORCE_SSL_ADMIN') ?: false);

// Set memory limits conservatively
@ini_set('memory_limit', '256M');

// Environment-specific settings (environment files own WP_DEBUG)
if (!in_array(WP_ENV, ['development', 'staging', 'production'], true)) {
    throw new \RuntimeException('Invalid WP_ENV value: ' . WP_ENV);
}
if (file_exists(__DIR__.'/environments/'.WP_ENV.'.php')) {
    require __DIR__.'/environments/'.WP_ENV.'.php';
}

// Debug fallback if environment file did not define WP_DEBUG
if (!defined('WP_DEBUG')) {
    define('WP_DEBUG', defined('WP_ENV') && WP_ENV !== 'production');
}
define('SCRIPT_DEBUG', WP_DEBUG);

// Absolute path to WordPress directory
define('ABSPATH', dirname(__DIR__).'/web/wp/');
