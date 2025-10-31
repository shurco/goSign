-- +goose Up
-- +goose StatementBegin
-- Create test accounts
INSERT INTO "public"."account" ("id", "name", "timezone", "locale", "created_at", "updated_at") VALUES 
('19ecfd4a-caf1-4ac9-91d7-21973fc9de31', 'Admin Account', 'UTC', 'en-US', NOW(), NOW()),
('375507c0-2d39-4d80-915a-6e89522915a7', 'User1 Account', 'UTC', 'en-US', NOW(), NOW()),
('c53aed39-0f8e-4926-843e-84db4a48de5c', 'User2 Account', 'UTC', 'en-US', NOW(), NOW());

-- Create test users with hashed passwords
-- Roles: 1 = User, 2 = Moderator, 3 = Admin
INSERT INTO "public"."user" ("id", "first_name", "last_name", "email", "role", "password", "account_id", "email_verified", "email_verified_at", "otp_enabled", "current_sign_in_at", "last_sign_in_at", "current_sign_in_ip", "last_sign_in_ip", "created_at", "updated_at") VALUES 
-- Administrator (role 3)
('ebf1ee29-ef5a-4aa9-8e7a-121fbcfc90bc', 'Admin', 'User', 'admin@gosign.local', 3, '$2a$04$OMbrej2usZtu4/XD1sCIg.JlctLQN54LzYBEW9sk72Hw75ikEOf2W', '19ecfd4a-caf1-4ac9-91d7-21973fc9de31', true, NOW(), false, NOW(), NOW(), '127.0.0.1', '127.0.0.1', NOW(), NOW()),
-- Regular User 1 (role 1)
('ef3a3b04-4d81-40a7-a387-cc572f68e23d', 'User', 'One', 'user1@gosign.local', 1, '$2a$04$OBX2Zs7g3N4Xk4RpHkgJEuYXhUg08UfwphA3SmUgqqxKB0UDEyLZi', '375507c0-2d39-4d80-915a-6e89522915a7', true, NOW(), false, NOW(), NOW(), '127.0.0.1', '127.0.0.1', NOW(), NOW()),
-- Regular User 2 (role 1)
('b57349ba-8ce0-4606-a87b-c20a2848a0b2', 'User', 'Two', 'user2@gosign.local', 1, '$2a$04$IDL5W8KYQJjpx/XAAvWw7.Rc0ZFPNl8wonpJQ.5xKL6ViPERs0g12', 'c53aed39-0f8e-4926-843e-84db4a48de5c', true, NOW(), false, NOW(), NOW(), '127.0.0.1', '127.0.0.1', NOW(), NOW());

-- Create default template folders for test accounts
INSERT INTO "public"."template_folder" ("id", "name", "account_id", "created_at", "updated_at") VALUES 
('765761bc-e27c-49de-9f4f-c8463bee2eb6', 'Default', '19ecfd4a-caf1-4ac9-91d7-21973fc9de31', NOW(), NOW()),
('263eca4f-47c4-4504-a3a5-0fa4c788f14d', 'Default', '375507c0-2d39-4d80-915a-6e89522915a7', NOW(), NOW()),
('c240a05b-e481-4941-b758-c26ce058dead', 'Default', 'c53aed39-0f8e-4926-843e-84db4a48de5c', NOW(), NOW());
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM "public"."template_folder" WHERE id IN (
  '765761bc-e27c-49de-9f4f-c8463bee2eb6',
  '263eca4f-47c4-4504-a3a5-0fa4c788f14d',
  'c240a05b-e481-4941-b758-c26ce058dead'
);

DELETE FROM "public"."user" WHERE id IN (
  'ebf1ee29-ef5a-4aa9-8e7a-121fbcfc90bc',
  'ef3a3b04-4d81-40a7-a387-cc572f68e23d',
  'b57349ba-8ce0-4606-a87b-c20a2848a0b2'
);

DELETE FROM "public"."account" WHERE id IN (
  '19ecfd4a-caf1-4ac9-91d7-21973fc9de31',
  '375507c0-2d39-4d80-915a-6e89522915a7',
  'c53aed39-0f8e-4926-843e-84db4a48de5c'
);
-- +goose StatementEnd

