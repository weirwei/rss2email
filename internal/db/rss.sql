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
)

CREATE UNIQUE INDEX uni_email_subscription ON user_subscriptions(email, subscription_id, subscription_type)
CREATE INDEX idx_subscription ON user_subscriptions(subscription_id)