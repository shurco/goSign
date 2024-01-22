-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION fn_updated_at()
	RETURNS TRIGGER
	LANGUAGE plpgsql
	AS $function$
BEGIN
	new.updated_at = NOW();
	RETURN NEW;
END;
$function$;

CREATE TRIGGER "tg_updated_at" BEFORE UPDATE ON "public"."setting" FOR EACH ROW EXECUTE FUNCTION fn_updated_at();
CREATE TRIGGER "tg_updated_at" BEFORE UPDATE ON "public"."account" FOR EACH ROW EXECUTE FUNCTION fn_updated_at();
CREATE TRIGGER "tg_updated_at" BEFORE UPDATE ON "public"."user" FOR EACH ROW EXECUTE FUNCTION fn_updated_at();
CREATE TRIGGER "tg_updated_at" BEFORE UPDATE ON "public"."template_folder" FOR EACH ROW EXECUTE FUNCTION fn_updated_at();
CREATE TRIGGER "tg_updated_at" BEFORE UPDATE ON "public"."template" FOR EACH ROW EXECUTE FUNCTION fn_updated_at();
CREATE TRIGGER "tg_updated_at" BEFORE UPDATE ON "public"."submission" FOR EACH ROW EXECUTE FUNCTION fn_updated_at();
CREATE TRIGGER "tg_updated_at" BEFORE UPDATE ON "public"."submitter" FOR EACH ROW EXECUTE FUNCTION fn_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER "tg_updated_at" ON "submitter";
DROP TRIGGER "tg_updated_at" ON "submission";
DROP TRIGGER "tg_updated_at" ON "template";
DROP TRIGGER "tg_updated_at" ON "template_folder";
DROP TRIGGER "tg_updated_at" ON "user";
DROP TRIGGER "tg_updated_at" ON "account";
DROP TRIGGER "tg_updated_at" ON "setting";
DROP FUNCTION fn_updated_at() CASCADE;
-- +goose StatementEnd