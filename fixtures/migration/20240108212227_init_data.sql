-- +goose Up
-- +goose StatementBegin
INSERT INTO "public"."account" ("id", "name", "timezone", "locale", "created_at", "updated_at") VALUES 
('12a57fea-3870-4ea5-bfe6-76f0741da2b3', 'CompanyName', 'London', 'en-GB', NOW(), NOW());

INSERT INTO "public"."user" ("id", "first_name", "last_name", "email", "role", "password", "account_id", "current_sign_in_at", "last_sign_in_at", "current_sign_in_ip", "last_sign_in_ip", "created_at", "updated_at") VALUES 
('3a3a7b08-61de-4448-8440-35dc589c51db', 'Firsname', 'Lastname', 'user@mail.com', 3, 'password', '12a57fea-3870-4ea5-bfe6-76f0741da2b3', NOW(), NOW(), '127.0.0.1', '127.0.0.1', NOW(), NOW());

INSERT INTO "public"."template_folder" ("id", "name", "account_id", "created_at", "updated_at") VALUES 
('2ab2d0fd-f331-4e12-a837-0c4cb3376fe7', 'Default', '12a57fea-3870-4ea5-bfe6-76f0741da2b3', NOW(), NOW());

INSERT INTO "public"."template" ("id", "folder_id", "slug", "name", "schema", "fields", "submitters", "source", "created_at", "updated_at") VALUES 
('00c95859-98ef-42cd-a801-2023b75a9431', '2ab2d0fd-f331-4e12-a837-0c4cb3376fe7', 'YhWszuL56FohQt', 'Example Template', 
'[{"attachment_id":"5a18b85d-e581-462c-903d-8ed00432fd39","name":"test1"},{"attachment_id":"62ce3aa9-145e-4d56-a58e-b0b0e1afa589","name":"test2"},{"attachment_id":"82255e6c-36f2-4518-b5e6-ed398180842f","name":"test3"}]', 
'[{"id":"4a3955e5-72bc-4f54-8670-7f1b751c7faa","submitter_id":"218c9a2e-be00-491d-bac6-056274eacc72","name":"Text field","type":"text","required":true,"preferences":{},"areas":[{"x":0.06818181818181818,"y":0.4016252642706131,"w":0.2522727272727273,"h":0.1564482029598309,"attachment_id":"62ce3aa9-145e-4d56-a58e-b0b0e1afa589","page":0}]},{"id":"34874b46-5b2d-443a-a917-47277b2ccaf8","submitter_id":"218c9a2e-be00-491d-bac6-056274eacc72","name":"First signer","type":"signature","required":true,"preferences":{},"areas":[{"x":0.07272727272727272,"y":0.1352404862579281,"w":0.2,"h":0.186046511627907,"attachment_id":"5a18b85d-e581-462c-903d-8ed00432fd39","page":0}]},{"id":"b5167ed9-bbf5-4756-bd30-0f9b823ed592","submitter_id":"ed038c87-efac-44ac-983f-3472c3960026","name":"Second signer","type":"signature","required":true,"preferences":{},"areas":[{"x":0.3931818181818182,"y":0.1204743657505285,"w":0.2,"h":0.186046511627907,"attachment_id":"5a18b85d-e581-462c-903d-8ed00432fd39","page":0}]},{"id":"d4036f37-0ed9-426f-ae87-21f9409ca92a","submitter_id":"46e3fc52-77bd-4794-92eb-7b515881a146","name":"Third signer","type":"signature","required":true,"preferences":{},"areas":[{"x":0.6818181818181818,"y":0.1141318710359408,"w":0.2,"h":0.186046511627907,"attachment_id":"5a18b85d-e581-462c-903d-8ed00432fd39","page":0}]}]', 
'[{"name":"First Party","id":"218c9a2e-be00-491d-bac6-056274eacc72"},{"name":"Second Party","id":"ed038c87-efac-44ac-983f-3472c3960026"},{"name":"Third Party","id":"46e3fc52-77bd-4794-92eb-7b515881a146"}]', 
'native', NOW(), NOW());

INSERT INTO "public"."storage_blob" ("id", "filename", "content_type", "metadata", "byte_size") VALUES 
('7d4bf7f6-e7d1-4277-a609-1a3f0eeb2001', '0.jpg', 'image/jpeg', '{"analyzed":true,"identified":true,"width":1400,"height":1980}', 45269),
('0f944506-9ae3-42e6-afdb-f34e297298fb', '0.jpg', 'image/jpeg', '{"width": 1400, "height": 1980, "analyzed": true, "identified": true}', 45269),
('f23be699-8bef-414a-bc26-e9dba6e69433', '0.jpg', 'image/jpeg', '{"width": 1400, "height": 1980, "analyzed": true, "identified": true}', 45269);


INSERT INTO "public"."storage_attachment" ("id", "blob_id", "name", "record_type", "record_id", "service_name", "created_at") VALUES 
('5a18b85d-e581-462c-903d-8ed00432fd39', '7d4bf7f6-e7d1-4277-a609-1a3f0eeb2001', 'documents', 'Template', '00c95859-98ef-42cd-a801-2023b75a9431', 'disk', NOW()),
('62ce3aa9-145e-4d56-a58e-b0b0e1afa589', '0f944506-9ae3-42e6-afdb-f34e297298fb', 'documents', 'Template', '00c95859-98ef-42cd-a801-2023b75a9431', 'disk', NOW()),
('82255e6c-36f2-4518-b5e6-ed398180842f', 'f23be699-8bef-414a-bc26-e9dba6e69433', 'documents', 'Template', '00c95859-98ef-42cd-a801-2023b75a9431', 'disk', NOW());
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM "public"."storage_attachment";
DELETE FROM "public"."storage_blob";
DELETE FROM "public"."template";
DELETE FROM "public"."template_folder";
DELETE FROM "public"."user";
DELETE FROM "public"."account";
-- +goose StatementEnd