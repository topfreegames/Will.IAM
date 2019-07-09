CREATE OR REPLACE FUNCTION resource_hierarchies_matches(text) RETURNS SETOF TEXT AS $$
  DECLARE
    parts text[];
    aux text[];
  BEGIN
    parts := string_to_array($1, '::');
    RETURN NEXT '*';
    FOR i IN array_lower(parts, 1) .. array_upper(parts, 1) - 1 LOOP
      aux := array_append(aux, parts[i]);
      RETURN NEXT array_to_string(array_append(aux, '*'), '::');
    END LOOP;
    RETURN NEXT $1;
    RETURN;
  END
$$ LANGUAGE 'plpgsql';
