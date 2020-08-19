/*======================================================================*/
--  20191103.sql
--  :mode=pl-sql:tabSize=3:indentSize=3:
--  Fri Feb 15 15:58:26 PST 2019 @40 /Internet Time/
--  Purpose:
/*======================================================================*/

\set ON_ERROR_STOP on
set client_min_messages to 'warning';

-- \cd %%subdir%%

/*======================================================================*/
--  Prevent application of this patch; comment out when ready for use.
/*======================================================================*/

/*
do $$
begin
   raise exception 'This patch is not yet usable -- aborting';
end $$;
*/

/*======================================================================*/
--  Test the database version.
/*======================================================================*/
do $$
DECLARE
   vers       text;
   check_vers text:='0.0.1';
BEGIN
   SELECT value INTO vers FROM config WHERE key='version';
   IF vers!=check_vers THEN
      raise exception 'Version % was not expected -- aborting',vers
         USING hint = ' Expected version is '||check_vers;
	END IF;
END $$;


/*======================================================================*/
--  Rev database version.
/*======================================================================*/
UPDATE
   config
SET
   value='5.0.1'
WHERE
   key='version';
