CREATE TABLE `user_subscriptions` (
    `id` integer PRIMARY KEY AUTOINCREMENT,
    `email` text,
    `subscription_id` text,
    `subscription_type` text,
    `process` text,
    `process_type` text,
    `created_at` datetime,
    `updated_at` datetime,
    `deleted` integer
);

CREATE UNIQUE INDEX uni_email_subscription ON user_subscriptions(email, subscription_id, subscription_type);
CREATE INDEX idx_subscription ON user_subscriptions(subscription_id);

CREATE TABLE `feed_sources` (
    `id` integer PRIMARY KEY AUTOINCREMENT,
    `subscription_id` text,
    `name` text,
    `feed_url` text,
    `language` text,
    `content_field` text,
    `schedule_type` text,
    `cron_spec` text,
    `created_at` datetime,
    `updated_at` datetime,
    `deleted` integer
);

CREATE UNIQUE INDEX uni_feed_source_subscription ON feed_sources(subscription_id);
