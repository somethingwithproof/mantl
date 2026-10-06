<?php
// SPDX-License-Identifier: MIT
/**
 * Front to the WordPress application.
 */

define('WP_USE_THEMES', true);

require dirname(__DIR__) . '/vendor/autoload.php';
require dirname(__DIR__) . '/config/application.php';

require ABSPATH . 'wp-blog-header.php';
