CREATE TABLE "public"."teaches"("id" serial NOT NULL, "created_at" timestamptz NOT NULL DEFAULT now(), "updated_at" timestamptz NOT NULL DEFAULT now(), "instructor_id" integer NOT NULL, "subject_id" integer NOT NULL, PRIMARY KEY ("id") , FOREIGN KEY ("instructor_id") REFERENCES "public"."instructors"("id") ON UPDATE cascade ON DELETE cascade, FOREIGN KEY ("subject_id") REFERENCES "public"."subjects"("id") ON UPDATE cascade ON DELETE cascade, UNIQUE ("instructor_id", "subject_id"));
CREATE OR REPLACE FUNCTION "public"."set_current_timestamp_updated_at"()
RETURNS TRIGGER AS $$
DECLARE
  _new record;
BEGIN
  _new := NEW;
  _new."updated_at" = NOW();
  RETURN _new;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER "set_public_teaches_updated_at"
BEFORE UPDATE ON "public"."teaches"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_current_timestamp_updated_at"();
COMMENT ON TRIGGER "set_public_teaches_updated_at" ON "public"."teaches" 
IS 'trigger to set value of column "updated_at" to current timestamp on row update';
