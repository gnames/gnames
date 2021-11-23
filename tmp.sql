with n as (
  select distinct v.name_string_id from word_name_strings wc
    join words w on w.id = wc.word_id
    join canonicals c on c.id = wc.canonical_id
    join verification v on c.id = v.canonical_id
    where modified like 'gallop%'
    and type_id = 12
    and c.name like 'M%'
    and v.classification like '%Coleoptera%'
),
au as (
  select distinct wc.name_string_id from word_name_strings wc
    join words w on w.id = wc.word_id
    join n on n.name_string_id = wc.name_string_id
    where w.modified = 'pic'
    and w.type_id = 4
)
select distinct au.name_string_id, v.name from verification v
 right join au on v.name_string_id = au.name_string_id
;
