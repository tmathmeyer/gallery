.open source.sqlite
ALTER TABLE Gallery_table add column Owner INTEGER;
ALTER TABLE Gallery_table add column Is_protected INTEGER;
ALTER TABLE Gallery_table add column Whitelist TEXT;
ALTER TABLE Gallery_table add column Splashwidth INTEGER;
ALTER TABLE Gallery_table add column Splashheight INTEGER;
UPDATE Gallery_table SET Splashwidth=1;
UPDATE Gallery_table SET Splashheight=1;
UPDATE Gallery_table SET Is_protected=0;
UPDATE Gallery_table SET Owner=0;
UPDATE Gallery_table SET Whitelist="";

ALTER TABLE Photo_table add column Width INTEGER;
ALTER TABLE Photo_table add column Height INTEGER;
UPDATE Photo_table SET Width=1;
UPDATE Photo_table SET Height=1;